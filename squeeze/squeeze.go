package squeeze

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type PlayerId string

func New(address string) Server {
	return Server{address: address, chanNotify: make(chan *Track)}
}

type Server struct {
	address    string
	chanNotify chan *Track
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

	err = s.fillArtist(conn, lms, id, &t)
	if err != nil {
		log.Warnf("unable to read artist field: %v", err)
	}

	err = s.fillAlbum(conn, lms, id, &t)
	if err != nil {
		log.Warnf("unable to read album field: %v", err)
	}

	err = s.fillTitle(conn, lms, id, &t)
	if err != nil {
		log.Warnf("unable to read title field: %v", err)
	}

	err = s.fillGenre(conn, lms, id, &t)
	if err != nil {
		log.Warnf("unable to read genre field: %v", err)
	}

	err = s.fillYear(conn, lms, id, &t)
	if err != nil {
		log.Warnf("invalid year field: %v", err)
	}

	err = s.fillDuration(conn, lms, id, &t)
	if err != nil {
		log.Warnf("unable to read duration field: %v", err)
	}

	err = s.fillTime(conn, lms, id, &t)
	if err != nil {
		log.Warnf("unable to read time field: %v", err)
	}

	return &t, nil
}

func (s *Server) currentTitle(writer io.Writer, reader *bufio.Reader, id PlayerId) (string, error) {
	_, err := fmt.Fprintf(writer, "%s current_title ?\r\n", id)
	if err != nil {
		return "", fmt.Errorf("unable to fetch track title: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("unable to read response: %v", err)
	}
	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	rawValue := strings.Split(line, " ")[2]
	title, err := url.QueryUnescape(rawValue)
	if err != nil {
		return "", fmt.Errorf("unable to unescape title \"%v\": %v", rawValue, err)
	}
	return title, nil
}

func (s *Server) fillArtist(writer io.Writer, reader *bufio.Reader, id PlayerId, t *Track) error {
	_, err := fmt.Fprintf(writer, "%s artist ?\r\n", id)
	if err != nil {
		return fmt.Errorf("unable to fetch track artist: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("unable to read response: %v", err)
	}

	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	rawValues := strings.Split(line, " ")
	if len(rawValues) < 3 {
		log.Debugf("no artist metadata for current track")
		return nil
	}
	rawValue := rawValues[2]
	artist, err := url.QueryUnescape(rawValue)
	if err != nil {
		return fmt.Errorf("unable to unescape artist \"%v\": %v", rawValue, err)
	}

	t.Artist = artist
	return nil
}

func (s *Server) fillAlbum(writer io.Writer, reader *bufio.Reader, id PlayerId, t *Track) error {
	_, err := fmt.Fprintf(writer, "%s album ?\r\n", id)
	if err != nil {
		return fmt.Errorf("unable to fetch track album: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("unable to read response: %v", err)
	}
	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	log.Debugf("parse raw album value '%v'", line)
	rawValues := strings.Split(line, " ")
	if len(rawValues) < 3 {
		log.Debugf("no album metadata for current track")
		return nil
	}
	rawValue := rawValues[2]
	album, err := url.QueryUnescape(rawValue)
	if err != nil {
		return fmt.Errorf("unable to unescape album \"%v\": %v", rawValue, err)
	}

	currentTitle, err := s.currentTitle(writer, reader, id)
	if err != nil {
		log.Warnf("unable to read current_title field: %v", err)
	}
	if strings.Index(currentTitle, "fip") == 0 || strings.Index(currentTitle, "FIP") == 0 {
		log.Debugf("search year in album '%v'", album)
		fields := strings.Split(album, "/")
		if len(fields) >= 2 {
			year, err := strconv.Atoi(strings.TrimSpace(fields[len(fields)-1]))
			if err != nil {
				log.Warnf("unable to parse year in album value \"%v\": %v", album, err)
			} else {
				log.Debugf("find year value '%v'", year)
				t.Year = year
			}
			album = strings.TrimSpace(strings.Join(fields[:len(fields)-1], "/"))
		}
	}
	log.Debugf("find album '%v'", album)
	t.Album = album
	return nil
}

func (s *Server) fillTitle(writer io.Writer, reader *bufio.Reader, id PlayerId, t *Track) error {
	_, err := fmt.Fprintf(writer, "%s title ?\r\n", id)
	if err != nil {
		return fmt.Errorf("unable to fetch track title: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("unable to read response: %v", err)
	}
	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	rawValue := strings.Split(line, " ")[2]
	title, err := url.QueryUnescape(rawValue)
	if err != nil {
		return fmt.Errorf("unable to unescape title \"%v\": %v", rawValue, err)
	}

	t.Title = title
	return nil
}

func (s *Server) fillGenre(writer io.Writer, reader *bufio.Reader, id PlayerId, t *Track) error {
	_, err := fmt.Fprintf(writer, "%s genre ?\r\n", id)
	if err != nil {
		return fmt.Errorf("unable to fetch track genre: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("unable to read response: %v", err)
	}
	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	rawValues := strings.Split(line, " ")
	if len(rawValues) < 3 {
		log.Debugf("no genre metadata for current track")
		return nil
	}
	rawValue := rawValues[2]
	genre, err := url.QueryUnescape(rawValue)
	if err != nil {
		return fmt.Errorf("unable to unescape genre \"%v\": %v", rawValue, err)
	}

	t.Genre = genre
	return nil
}

func (s *Server) fillYear(writer io.Writer, reader *bufio.Reader, id PlayerId, t *Track) error {
	_, err := fmt.Fprintf(writer, "%s year ?\r\n", id)
	if err != nil {
		return fmt.Errorf("unable to fetch track year: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("unable to read response: %v", err)
	}

	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	rawValue := strings.Split(line, " ")[2]
	rawYear, err := url.QueryUnescape(rawValue)
	if err != nil {
		return fmt.Errorf("unable to unescape year \"%v\": %v", rawValue, err)
	}

	year, err := strconv.Atoi(strings.Trim(rawYear, "\r"))
	if err != nil {
		return fmt.Errorf("unable to parse year value \"%v\": %v", rawLine, err)
	}
	t.Year = year
	return nil
}

func (s *Server) fillDuration(writer io.Writer, reader *bufio.Reader, id PlayerId, t *Track) error {
	_, err := fmt.Fprintf(writer, "%s duration ?\r\n", id)
	if err != nil {
		return fmt.Errorf("unable to fetch track duration: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("unable to read response: %v", err)
	}
	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	rawValue := strings.Split(line, " ")[2]

	d, err := strconv.ParseFloat(strings.Trim(rawValue, "\r"), 64)
	if err != nil {
		return fmt.Errorf("unable to parse duration value \"%v\": %v", rawLine, err)
	}
	t.Duration = d
	return nil
}

func (s *Server) fillTime(writer io.Writer, reader *bufio.Reader, id PlayerId, t *Track) error {
	_, err := fmt.Fprintf(writer, "%s time ?\r\n", id)
	if err != nil {
		return fmt.Errorf("unable to fetch track current time: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("unable to read response: %v", err)
	}

	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	rawValue := strings.Split(line, " ")[2]

	tm, err := strconv.ParseFloat(strings.Trim(rawValue, "\r"), 64)
	if err != nil {
		return fmt.Errorf("unable to parse time value \"%v\": %v", rawLine, err)
	}
	t.CurrentTime = tm
	return nil
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
