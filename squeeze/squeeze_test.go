package squeeze

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
)

const (
	playerId = PlayerId("playerId")
)

func Test_ParseAlbum(t *testing.T) {
	log.SetLevel(log.InfoLevel)
	cases := []struct {
		name            string
		rawCurrentTitle string
		rawAlbum        string
		expectedAlbum   string
		expectedYear    int
	}{
		{"Generic", "other", "Innervisions%20%2F%201973", "Innervisions / 1973", 0},
		{"On RadioFrance", "FIP", "Innervisions%20%2F%201973", "Innervisions", 1973},
		{"Separator on album name", "FIP", "BOF%20%2F%20The%20irishman%20%2F%201959", "BOF / The irishman", 1959},
	}
	squeezeMock := ConnMock{}
	err := squeezeMock.listen()
	if err != nil {
		t.Errorf("unable to start mock squeeze server: %v", err)
	}
	defer squeezeMock.Close()

	server := New(squeezeMock.Addr())
	for _, c := range cases {
		squeezeMock.SetRawTrack(RawTrack{rawCurrentTitle: c.rawCurrentTitle, rawAlbum: c.rawAlbum})

		track, err := server.CurrentTrack(playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if track.Album != c.expectedAlbum {
			t.Errorf("[%v] bad album: %#v, wants %#v", c.name, track.Album, c.expectedAlbum)
		}
		if track.Year != c.expectedYear {
			t.Errorf("[%v] bad year: %#v, wants %#v", c.name, track.Year, c.expectedYear)
		}
	}

}
func TestServer_CurrentTrack(t *testing.T) {

	cases := []struct {
		name          string
		rawTrack      RawTrack
		expectedTrack Track
	}{
		{"Net Radio 1",
			RawTrack{
				"fipelectro-midfi.mp3",
				"Tenderlonious",
				"On%20flute%20%2F%202019",
				"In%20A%20Sentimental%20Mood", "",
				"%3f", 272, 866},
			Track{
				Artist:      "Tenderlonious",
				Album:       "On flute",
				Title:       "In A Sentimental Mood",
				Genre:       "",
				Year:        2019,
				CurrentTime: 272,
				Duration:    866,
			}},
		{"Net Radio 2",
			RawTrack{
				"FIP",
				"Little%20Richard",
				"Little%20Richard%20%2F%201956",
				"I%20brought%20it%20all%20on%20myself",
				"",
				"%3F", 171, 33.761625246048},
			Track{
				Artist:      "Little Richard",
				Album:       "Little Richard",
				Title:       "I brought it all on myself",
				Genre:       "",
				Year:        1956,
				CurrentTime: 171,
				Duration:    33.761625246048,
			},
		},
	}

	squeezeMock := ConnMock{}
	err := squeezeMock.listen()
	if err != nil {
		t.Errorf("unable to start mock squeeze server: %v", err)
	}
	defer squeezeMock.Close()

	server := New(squeezeMock.Addr())
	for _, c := range cases {
		squeezeMock.SetRawTrack(c.rawTrack)

		track, err := server.CurrentTrack("player-id")
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if *track != c.expectedTrack {
			t.Errorf("[%v] bad track: %#v, wants %#v", c.name, *track, c.expectedTrack)
		}
	}
}

type RawTrack struct {
	rawCurrentTitle string
	rawArtist       string
	rawAlbum        string
	rawTitle        string
	rawGenre        string
	rawYear         string
	rawCurrentTime  float64
	rawDuration     float64
}

type ConnMock struct {
	muTrack sync.Mutex
	track   RawTrack

	ln net.Listener
}

func (c *ConnMock) listen() error {
	ln, err := net.Listen("tcp", "127.0.0.1:")
	c.ln = ln
	if err != nil {

		return fmt.Errorf("unable to listen on port: %v", err)
	}

	go func() {
		for {
			conn, err := c.ln.Accept()
			if err != nil {
				log.Infof("connection close: %v", err)
				break
			}
			go c.handleConnection(conn)
		}
	}()
	return nil
}

func (c *ConnMock) Addr() string {
	return c.ln.Addr().String()
}

func (c *ConnMock) SetRawTrack(track RawTrack) {
	c.muTrack.Lock()
	defer c.muTrack.Unlock()
	c.track = track
}

func (c *ConnMock) handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	for {

		rawCmd, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Info("connection closed")
				break
			}
			log.Errorf("unable to read request: %v", err)
		}
		args := strings.Split(rawCmd, " ")
		player := PlayerId(args[0])
		action := args[1]

		err = c.writeResponse(action, writer, player)
		if err != nil {
			log.Errorf("unable to write response: %v", err)
		}
		if err == io.EOF {
			log.Info("Connection closed")
			break
		}
	}
}

func (c *ConnMock) writeResponse(action string, writer *bufio.Writer, player PlayerId) error {
	c.muTrack.Lock()
	defer c.muTrack.Unlock()
	log.Debugf("action: %v", action)
	var err error
	switch action {
	case "artist":
		_, err = writer.WriteString(fmt.Sprintf("%v %v %v\r\n", player, action, c.track.rawArtist))
	case "current_title":
		_, err = writer.WriteString(fmt.Sprintf("%v %v %v\r\n", player, action, c.track.rawCurrentTitle))
	case "album":
		_, err = writer.WriteString(fmt.Sprintf("%v %v %v\r\n", player, action, c.track.rawAlbum))
	case "title":
		_, err = writer.WriteString(fmt.Sprintf("%v %v %v\r\n", player, action, c.track.rawTitle))
	case "genre":
		_, err = writer.WriteString(fmt.Sprintf("%v %v %v\r\n", player, action, c.track.rawGenre))
	case "year":
		_, err = writer.WriteString(fmt.Sprintf("%v %v %v\r\n", player, action, c.track.rawYear))
	case "time":
		_, err = writer.WriteString(fmt.Sprintf("%v %v %v\r\n", player, action, c.track.rawCurrentTime))
	case "duration":
		_, err = writer.WriteString(fmt.Sprintf("%v %v %v\r\n", player, action, c.track.rawDuration))
	default:
		_, err = writer.WriteString(fmt.Sprintf("%v %v %v\r\n", player, action, ""))
	}
	if err != nil {
		return fmt.Errorf("unableto write response: %v", err)
	}
	return writer.Flush()
}

func (c *ConnMock) Close() error {
	log.Debug("close mock server")
	err := c.ln.Close()
	if err != nil {
		return fmt.Errorf("unable to close mock server: %v", err)
	}
	return nil
}
