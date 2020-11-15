package squeeze

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/url"
	"strconv"
	"strings"
)

type MetadataParser interface {
	ArtistParser
	AlbumParser
	TitleParser
	YearParser
	GenreParser
	DurationParser
	TimeParser
}

type CurrentTitleParser interface {
	CurrentTitle(writer io.Writer, reader *bufio.Reader, id PlayerId) (string, error)
}

type ArtistParser interface {
	Artist(writer io.Writer, reader *bufio.Reader, id PlayerId) (string, error)
}

type AlbumParser interface {
	Album(writer io.Writer, reader *bufio.Reader, id PlayerId) (string, error)
}

type TitleParser interface {
	Title(writer io.Writer, reader *bufio.Reader, id PlayerId) (string, error)
}

type YearParser interface {
	Year(writer io.Writer, reader *bufio.Reader, id PlayerId) (int, error)
}

type GenreParser interface {
	Genre(writer io.Writer, reader *bufio.Reader, id PlayerId) (string, error)
}

type TrackDuration float64

var NilTrackDuration = TrackDuration(-1)

type DurationParser interface {
	Duration(writer io.Writer, reader *bufio.Reader, id PlayerId) (TrackDuration, error)
}

type TrackTime float64

var NilTrackTime = TrackTime(-1)

type TimeParser interface {
	Time(writer io.Writer, reader *bufio.Reader, id PlayerId) (TrackTime, error)
}

type DefaultParser struct {
	DefaultArtistParser
	DefaultAlbumParser
	DefaultTitleParser
	DefaultYearParser
	DefaultGenreParser
	DefaultDurationParser
	DefaultTimeParser
}

type DefaultArtistParser struct{}

func (p DefaultArtistParser) Artist(writer io.Writer, reader *bufio.Reader, id PlayerId) (string, error) {
	_, err := fmt.Fprintf(writer, "%s artist ?\r\n", id)
	if err != nil {
		return "", fmt.Errorf("unable to fetch track artist: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("unable to read response: %v", err)
	}

	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	rawValues := strings.Split(line, " ")
	if len(rawValues) < 3 {
		log.Debugf("no artist metadata for current track")
		return "", nil
	}
	rawValue := rawValues[2]
	artist, err := url.QueryUnescape(rawValue)
	if err != nil {
		return "", fmt.Errorf("unable to unescape artist \"%v\": %v", rawValue, err)
	}

	return artist, nil
}

type DefaultAlbumParser struct{}

func (p DefaultAlbumParser) Album(writer io.Writer, reader *bufio.Reader, id PlayerId) (string, error) {
	_, err := fmt.Fprintf(writer, "%s album ?\r\n", id)
	if err != nil {
		return "", fmt.Errorf("unable to fetch track album: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("unable to read response: %v", err)
	}
	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	log.Debugf("parse raw album value '%v'", line)
	rawValues := strings.Split(line, " ")
	if len(rawValues) < 3 {
		log.Debugf("no album metadata for current track")
		return "", nil
	}
	rawValue := rawValues[2]
	album, err := url.QueryUnescape(rawValue)
	if err != nil {
		return "", fmt.Errorf("unable to unescape album \"%v\": %v", rawValue, err)
	}
	log.Debugf("find album '%v'", album)
	return album, nil
}

type DefaultYearParser struct{}

func (p DefaultYearParser) Year(writer io.Writer, reader *bufio.Reader, id PlayerId) (int, error) {
	_, err := fmt.Fprintf(writer, "%s year ?\r\n", id)
	if err != nil {
		return 0, fmt.Errorf("unable to fetch track year: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("unable to read response: %v", err)
	}

	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	rawValue := strings.Split(line, " ")[2]
	rawYear, err := url.QueryUnescape(rawValue)
	if err != nil {
		return 0, fmt.Errorf("unable to unescape year \"%v\": %v", rawValue, err)
	}

	year, err := strconv.Atoi(strings.Trim(rawYear, "\r"))
	if err != nil {
		log.WithFields(log.Fields{"parser": "year", "rawYear": rawYear}).Debug("year field isn't an integer, value ignored")
		year = 0
	}
	return year, nil
}

type DefaultCurrentTitleParser struct{}

func (p DefaultCurrentTitleParser) CurrentTitle(writer io.Writer, reader *bufio.Reader, id PlayerId) (string, error) {
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

type DefaultTitleParser struct{}

func (p DefaultTitleParser) Title(writer io.Writer, reader *bufio.Reader, id PlayerId) (string, error) {
	_, err := fmt.Fprintf(writer, "%s title ?\r\n", id)
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

type DefaultGenreParser struct{}

func (p DefaultGenreParser) Genre(writer io.Writer, reader *bufio.Reader, id PlayerId) (string, error) {
	_, err := fmt.Fprintf(writer, "%s genre ?\r\n", id)
	if err != nil {
		return "", fmt.Errorf("unable to fetch track genre: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("unable to read response: %v", err)
	}
	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	rawValues := strings.Split(line, " ")
	if len(rawValues) < 3 {
		log.Debugf("no genre metadata for current track")
		return "", nil
	}
	rawValue := rawValues[2]
	genre, err := url.QueryUnescape(rawValue)
	if err != nil {
		return "", fmt.Errorf("unable to unescape genre \"%v\": %v", rawValue, err)
	}
	return genre, nil
}

type DefaultDurationParser struct{}

func (p DefaultDurationParser) Duration(writer io.Writer, reader *bufio.Reader, id PlayerId) (TrackDuration, error) {
	_, err := fmt.Fprintf(writer, "%s duration ?\r\n", id)
	if err != nil {
		return NilTrackDuration, fmt.Errorf("unable to fetch track duration: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return NilTrackDuration, fmt.Errorf("unable to read response: %v", err)
	}
	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	rawValue := strings.Split(line, " ")[2]

	d, err := strconv.ParseFloat(strings.Trim(rawValue, "\r"), 64)
	if err != nil {
		return NilTrackDuration, fmt.Errorf("unable to parse duration value \"%v\": %v", rawLine, err)
	}
	return TrackDuration(d), nil
}

type DefaultTimeParser struct{}

func (p DefaultTimeParser) Time(writer io.Writer, reader *bufio.Reader, id PlayerId) (TrackTime, error) {
	_, err := fmt.Fprintf(writer, "%s time ?\r\n", id)
	if err != nil {
		return NilTrackTime, fmt.Errorf("unable to fetch track current time: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return NilTrackTime, fmt.Errorf("unable to read response: %v", err)
	}

	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	rawValue := strings.Split(line, " ")[2]

	tm, err := strconv.ParseFloat(strings.Trim(rawValue, "\r"), 64)
	if err != nil {
		return NilTrackTime, fmt.Errorf("unable to parse time value \"%v\": %v", rawLine, err)
	}
	return TrackTime(tm), nil
}

type RadioFranceParser struct {
	yearParser DefaultYearParser
	DefaultTitleParser
	DefaultArtistParser
	DefaultGenreParser
	DefaultDurationParser
	DefaultTimeParser
}

func (r RadioFranceParser) Year(writer io.Writer, reader *bufio.Reader, id PlayerId) (int, error) {
	line, err := r.readAlbumMetadata(writer, reader, id)
	if err != nil {
		return 0, fmt.Errorf("unable to parse result to read album: %v", err)
	}

	log.Debugf("search year in album metadata '%v'", line)
	fields := strings.Split(line, "/")
	if len(fields) >= 2 {
		year, err := strconv.Atoi(strings.TrimSpace(fields[len(fields)-1]))
		if err != nil {
			log.Warnf("unable to parse year in album value \"%v\": %v", line, err)
		} else {
			log.Debugf("find year value '%v'", year)
			return year, nil
		}
	}
	log.Debug("no year found in album line, search it in default fields")
	return r.yearParser.Year(writer, reader, id)
}

func (r RadioFranceParser) Album(writer io.Writer, reader *bufio.Reader, id PlayerId) (string, error) {

	var album string
	line, err := r.readAlbumMetadata(writer, reader, id)
	if err != nil {
		return "", fmt.Errorf("unable to parse result to read album: %v", err)
	}

	log.Debugf("check if year exists in album metadata '%v'", line)
	fields := strings.Split(line, "/")
	if len(fields) >= 2 {
		album = strings.TrimSpace(strings.Join(fields[:len(fields)-1], "/"))
	} else {
		album = strings.TrimSpace(line)
	}

	log.Debugf("find album '%v'", album)
	return album, nil
}

func (r RadioFranceParser) readAlbumMetadata(writer io.Writer, reader *bufio.Reader, id PlayerId) (string, error) {
	_, err := fmt.Fprintf(writer, "%s album ?\r\n", id)
	if err != nil {
		return "", fmt.Errorf("unable to fetch track escapedLine: %v", err)
	}

	rawLine, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("unable to read response: %v", err)
	}
	line := strings.ReplaceAll(rawLine, "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	log.Debugf("parse raw escapedLine value '%v'", line)
	rawValues := strings.Split(line, " ")
	if len(rawValues) < 3 {
		log.Debugf("no escapedLine metadata for current track")
		return "", nil
	}
	rawValue := rawValues[2]
	escapedLine, err := url.QueryUnescape(rawValue)
	if err != nil {
		return "", fmt.Errorf("unable to unescape escapedLine \"%v\": %v", rawValue, err)
	}
	return escapedLine, nil
}
