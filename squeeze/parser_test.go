package squeeze

import (
	"bufio"
	"io"
	"testing"
)
import log "github.com/sirupsen/logrus"

func TestParser(t *testing.T) {

	squeezeMock := ConnMock{}
	err := squeezeMock.listen()
	if err != nil {
		t.Errorf("unable to start mock squeeze server: %v", err)
	}
	defer squeezeMock.Close()

	conn, err := connect(squeezeMock.Addr())
	if err != nil {
		t.Errorf("unable to connect to '%v' server: %v", squeezeMock.Addr(), err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Warnf("unable to close connection to lms server after track parsing: %v", err)
		}
	}()

	log.SetLevel(log.InfoLevel)

	lms := bufio.NewReader(conn)

	t.Run("CurrentTitle", func(t *testing.T) { DefaultParserCurrentTitle(t, &squeezeMock, conn, lms) })

	t.Run("Album", func(t *testing.T) { DefaultParserAlbum(t, &squeezeMock, conn, lms) })
	t.Run("Title", func(t *testing.T) { DefaultParserTitle(t, &squeezeMock, conn, lms) })
	t.Run("Artist", func(t *testing.T) { DefaultParserArtist(t, &squeezeMock, conn, lms) })
	t.Run("Year", func(t *testing.T) { DefaultParserYear(t, &squeezeMock, conn, lms) })
	t.Run("Genre", func(t *testing.T) { DefaultParserGenre(t, &squeezeMock, conn, lms) })
	t.Run("Duration", func(t *testing.T) { DefaultParserDuration(t, &squeezeMock, conn, lms) })
	t.Run("Time", func(t *testing.T) { DefaultParserTime(t, &squeezeMock, conn, lms) })

	t.Run("RadioFranceAlbum", func(t *testing.T) { RadioFranceParserAlbum(t, &squeezeMock, conn, lms) })
	t.Run("RadioFranceTitle", func(t *testing.T) { RadioFranceParserTitle(t, &squeezeMock, conn, lms) })
	t.Run("RadioFranceArtist", func(t *testing.T) { RadioFranceParserArtist(t, &squeezeMock, conn, lms) })
	t.Run("RadioFranceYear", func(t *testing.T) { RadioFranceParserYear(t, &squeezeMock, conn, lms) })
	t.Run("RadioFranceGenre", func(t *testing.T) { RadioFranceParserGenre(t, &squeezeMock, conn, lms) })
	t.Run("RadioFranceDuration", func(t *testing.T) { RadioFranceParserDuration(t, &squeezeMock, conn, lms) })
	t.Run("RadioFranceTime", func(t *testing.T) { RadioFranceParserTime(t, &squeezeMock, conn, lms) })
}

func DefaultParserCurrentTitle(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name                 string
		rawCurrentTitle      string
		expectedCurrentTitle string
	}{
		{"Simple", "fipelectro-midfi.mp3", "fipelectro-midfi.mp3"},
		{"Not defined", "", ""},
	}

	parser := DefaultCurrentTitleParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawCurrentTitle: c.rawCurrentTitle})

		title, err := parser.CurrentTitle(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if title != c.expectedCurrentTitle {
			t.Errorf("[%v] bad title: %#v, wants %#v", c.name, title, c.expectedCurrentTitle)
		}
	}
}

func DefaultParserAlbum(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name          string
		rawAlbum      string
		expectedAlbum string
	}{
		{"Simple", "Innervisions%20%2F%201973", "Innervisions / 1973"},
		{"With /", "Innervisions%20%2F%201973", "Innervisions / 1973"},
		{"Separator on name", "BOF%20%2F%20The%20irishman%20%2F%201959", "BOF / The irishman / 1959"},
		{"Not defined", "", ""},
	}

	parser := DefaultAlbumParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawAlbum: c.rawAlbum})

		album, err := parser.Album(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if album != c.expectedAlbum {
			t.Errorf("[%v] bad album: %#v, wants %#v", c.name, album, c.expectedAlbum)
		}
	}
}

func DefaultParserTitle(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name          string
		rawTitle      string
		expectedTitle string
	}{
		{"Simple", "I%20brought%20it%20all%20on%20myself", "I brought it all on myself"},
		{"Not defined", "", ""},
	}

	parser := DefaultTitleParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawTitle: c.rawTitle})

		title, err := parser.Title(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if title != c.expectedTitle {
			t.Errorf("[%v] bad title: %#v, wants %#v", c.name, title, c.expectedTitle)
		}
	}
}

func DefaultParserArtist(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name           string
		rawArtist      string
		expectedArtist string
	}{
		{"Simple", "Little%20Richard", "Little Richard"},
		{"Not defined", "", ""},
	}

	parser := DefaultArtistParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawArtist: c.rawArtist})

		value, err := parser.Artist(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if value != c.expectedArtist {
			t.Errorf("[%v] bad artist: %#v, wants %#v", c.name, value, c.expectedArtist)
		}
	}
}

func DefaultParserYear(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name         string
		rawYear      string
		expectedYear int
	}{
		{"Simle", "2018", 2018},
		{"Year unknown", "%3f", 0},
		{"Year unknown 2", "%3F", 0},
		{"Not defined", "", 0},
	}

	parser := DefaultYearParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawYear: c.rawYear})

		value, err := parser.Year(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if value != c.expectedYear {
			t.Errorf("[%v] bad year: %#v, wants %#v", c.name, value, c.expectedYear)
		}
	}
}

func DefaultParserGenre(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name          string
		rawGenre      string
		expectedGenre string
	}{
		{"Simple", "Jazz", "Jazz"},
		{"With encoded character", "An%20other%20genre", "An other genre"},
		{"Not defined", "", ""},
	}

	parser := DefaultGenreParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawGenre: c.rawGenre})

		value, err := parser.Genre(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if value != c.expectedGenre {
			t.Errorf("[%v] bad genre: %#v, wants %#v", c.name, value, c.expectedGenre)
		}
	}
}

func DefaultParserDuration(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name             string
		rawDuration      float64
		expectedDuration TrackDuration
	}{
		{"Simle", 100., 100},
		{"Not defined", -1, -1},
	}

	parser := DefaultDurationParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawDuration: c.rawDuration})

		value, err := parser.Duration(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if value != c.expectedDuration {
			t.Errorf("[%v] bad duration: %#v, wants %#v", c.name, value, c.expectedDuration)
		}
	}
}

func DefaultParserTime(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name         string
		rawTime      float64
		expectedTime TrackTime
	}{
		{"Simle", 100., 100},
		{"Not defined", -1, -1},
	}

	parser := DefaultTimeParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawCurrentTime: c.rawTime})

		value, err := parser.Time(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if value != c.expectedTime {
			t.Errorf("[%v] bad time: %#v, wants %#v", c.name, value, c.expectedTime)
		}
	}
}

func RadioFranceParserAlbum(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name          string
		rawAlbum      string
		expectedAlbum string
	}{
		{"Simple", "Innervisions", "Innervisions"},
		{"With /", "Innervisions%20%2F%201973", "Innervisions"},
		{"Separator on name", "BOF%20%2F%20The%20irishman%20%2F%201959", "BOF / The irishman"},
		{"Not defined", "", ""},
	}

	parser := RadioFranceParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawAlbum: c.rawAlbum})

		album, err := parser.Album(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if album != c.expectedAlbum {
			t.Errorf("[%v] bad album: %#v, wants %#v", c.name, album, c.expectedAlbum)
		}
	}
}

func RadioFranceParserTitle(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name          string
		rawTitle      string
		expectedTitle string
	}{
		{"Simple", "I%20brought%20it%20all%20on%20myself", "I brought it all on myself"},
		{"Not defined", "", ""},
	}

	parser := RadioFranceParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawTitle: c.rawTitle})

		title, err := parser.Title(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if title != c.expectedTitle {
			t.Errorf("[%v] bad title: %#v, wants %#v", c.name, title, c.expectedTitle)
		}
	}
}

func RadioFranceParserArtist(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name           string
		rawArtist      string
		expectedArtist string
	}{
		{"Simple", "Little%20Richard", "Little Richard"},
		{"Not defined", "", ""},
	}

	parser := RadioFranceParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawArtist: c.rawArtist})

		value, err := parser.Artist(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if value != c.expectedArtist {
			t.Errorf("[%v] bad artist: %#v, wants %#v", c.name, value, c.expectedArtist)
		}
	}
}

func RadioFranceParserYear(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name         string
		rawAlbum     string
		rawYear      string
		expectedYear int
	}{
		{"Simle", "", "2018", 2018},
		{"Year unknown", "", "%3f", 0},
		{"Year unknown 2", "", "%3F", 0},
		{"Not defined", "album", "", 0},
		{"Year in album field", "Innervisions%20%2F%201973", "", 1973},
		{"Year in album and year fields", "Innervisions%20%2F%201973", "1981", 1973},
	}

	parser := RadioFranceParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawYear: c.rawYear, rawAlbum: c.rawAlbum})

		value, err := parser.Year(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if value != c.expectedYear {
			t.Errorf("[%v] bad year: %#v, wants %#v", c.name, value, c.expectedYear)
		}
	}
}

func RadioFranceParserGenre(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name          string
		rawGenre      string
		expectedGenre string
	}{
		{"Simple", "Jazz", "Jazz"},
		{"With encoded character", "An%20other%20genre", "An other genre"},
		{"Not defined", "", ""},
	}

	parser := RadioFranceParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawGenre: c.rawGenre})

		value, err := parser.Genre(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if value != c.expectedGenre {
			t.Errorf("[%v] bad genre: %#v, wants %#v", c.name, value, c.expectedGenre)
		}
	}
}

func RadioFranceParserDuration(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name             string
		rawDuration      float64
		expectedDuration TrackDuration
	}{
		{"Simle", 100., 100},
		{"Not defined", -1, -1},
	}

	parser := RadioFranceParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawDuration: c.rawDuration})

		value, err := parser.Duration(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if value != c.expectedDuration {
			t.Errorf("[%v] bad duration: %#v, wants %#v", c.name, value, c.expectedDuration)
		}
	}
}

func RadioFranceParserTime(t *testing.T, mock *ConnMock, conn io.ReadWriteCloser, lms *bufio.Reader) {
	cases := []struct {
		name         string
		rawTime      float64
		expectedTime TrackTime
	}{
		{"Simle", 100., 100},
		{"Not defined", -1, -1},
	}

	parser := RadioFranceParser{}
	for _, c := range cases {
		mock.SetRawTrack(RawTrack{rawCurrentTime: c.rawTime})

		value, err := parser.Time(conn, lms, playerId)
		if err != nil {
			t.Errorf("[%v] unable to read track infos: %v", c.name, err)
		}
		if value != c.expectedTime {
			t.Errorf("[%v] bad time: %#v, wants %#v", c.name, value, c.expectedTime)
		}
	}
}
