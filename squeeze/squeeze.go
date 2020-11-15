package squeeze

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"strings"
)

type PlayerId string

func New(address string) Server {
	return Server{address: address, chanNotify: make(chan *Track)}
}

type Server struct {
	address    string
	chanNotify chan *Track
	DefaultCurrentTitleParser
}

func (s *Server) Close() error {
	close(s.chanNotify)
	return nil
}

func (s *Server) Listen() error {

	conn, err := connect(s.address)
	if err != nil {
		return fmt.Errorf("unable to connect to %v", s.address)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Warnf("unable to close connection to server %v: %v", s.address, err)
		}
	}()

	_, err = fmt.Fprintf(conn, "listen 1\n")
	if err != nil {
		return fmt.Errorf("unable to send 'listen' command to server: %v", err)
	}

	lms := bufio.NewReader(conn)
	for {
		rawLine, err := lms.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Debugf("connection to server close: %v", err)
				break
			}
			log.Errorf("unable to read response: %v", err)
			continue
		}
		line := strings.Trim(rawLine, "\r")
		s.processEventLine(line)
	}

	return nil
}

func (s *Server) processEventLine(line string) {
	log.Infof("new event: %v", line)
	if strings.Contains(line, " newmetadata\n") ||
		strings.Contains(line, " newsong ") {
		s.onNewMetadata(line)
	}
}

func (s *Server) NotifyTrackChange() <-chan *Track {
	return s.chanNotify
}

type Track struct {
	Artist      string
	Album       string
	Title       string
	Genre       string
	Year        int
	CurrentTime float64
	Duration    float64
}

func (s *Server) CurrentTrack(id PlayerId) (*Track, error) {
	conn, err := connect(s.address)
	defer func() {
		if err := conn.Close(); err != nil {
			log.Warnf("unable to close connection to lms server after track parsing: %v", err)
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to '%v' server: %v", s.address, err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Warnf("unable to close connection to '%v' server: %v", s.address, err)
		}
	}()

	lms := bufio.NewReader(conn)
	t := Track{}

	currentTitle, err := s.CurrentTitle(conn, lms, id)
	if err != nil {
		log.Warnf("unable to read current_title field: %v", err)
	}

	var mp MetadataParser
	if strings.Index(currentTitle, "fip") == 0 || strings.Index(currentTitle, "FIP") == 0 {
		mp = RadioFranceParser{}
	} else {
		mp = DefaultParser{}
	}

	artist, err := mp.Artist(conn, lms, id)
	if err != nil {
		log.Warnf("unable to read artist field: %v", err)
	} else {
		t.Artist = artist
	}

	album, err := mp.Album(conn, lms, id)
	if err != nil {
		log.Warnf("unable to read album field: %v", err)
	} else {
		t.Album = album
	}

	title, err := mp.Title(conn, lms, id)
	if err != nil {
		log.Warnf("unable to read title field: %v", err)
	} else {
		t.Title = title
	}

	genre, err := mp.Genre(conn, lms, id)
	if err != nil {
		log.Warnf("unable to read genre field: %v", err)
	} else {
		t.Genre = genre
	}

	year, err := mp.Year(conn, lms, id)
	if err != nil {
		log.Warnf("invalid year field: %v", err)
	} else if year > 0 {
		t.Year = year
	}

	d, err := mp.Duration(conn, lms, id)
	if err != nil {
		log.Warnf("unable to read duration field: %v", err)
	} else if d != NilTrackDuration {
		t.Duration = float64(d)
	}

	tm, err := mp.Time(conn, lms, id)
	if err != nil {
		log.Warnf("unable to read time field: %v", err)
	} else if tm != NilTrackTime {
		t.CurrentTime = float64(tm)
	}

	return &t, nil
}

func (s *Server) onNewMetadata(line string) {
	args := strings.Split(strings.Trim(line, "\n"), " ")
	id := PlayerId(args[0])
	t, err := s.CurrentTrack(id)
	if err != nil {
		log.Errorf("unable to extract current track metadata for player %v: %v", id, err)
	}
	s.chanNotify <- t
}

var connect = func(address string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to %v", address)
	}
	return conn, nil
}
