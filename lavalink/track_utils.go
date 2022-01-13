package lavalink

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io"
	"time"
)

const trackInfoVersioned int = 1
const trackInfoVersion int = 2

func EncodeToString(info TrackInfo) (str string, err error) {
	w := new(bytes.Buffer)

	if err = WriteInt32(w, int32(trackInfoVersion)); err != nil {
		return
	}
	if err = WriteString(w, info.Title()); err != nil {
		return
	}
	if err = WriteString(w, info.Author()); err != nil {
		return
	}
	if err = WriteInt64(w, info.Length().Milliseconds()); err != nil {
		return
	}
	if err = WriteString(w, info.Identifier()); err != nil {
		return
	}
	if err = WriteBool(w, info.IsStream()); err != nil {
		return
	}
	if err = WriteBool(w, info.URI() != nil); err != nil {
		return
	}
	if err = WriteNullableString(w, info.URI()); err != nil {
		return
	}
	if err = WriteString(w, info.SourceName()); err != nil {
		return
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, int32(w.Len()|trackInfoVersioned<<30))
	if err != nil {
		return
	}
	buf.Write(w.Bytes())

	str = base64.StdEncoding.EncodeToString(buf.Bytes())
	return
}

// DecodeString thx to https://github.com/foxbot/gavalink/blob/master/decoder.go
func DecodeString(str string) (info TrackInfo, err error) {

	var data []byte
	data, err = base64.StdEncoding.DecodeString(str)
	if err != nil {
		return
	}

	r := bytes.NewReader(data)

	trackInfo := &DefaultTrackInfo{}

	var value uint8
	if err = binary.Read(r, binary.LittleEndian, &value); err != nil {
		return
	}

	flags := int32(int64(value) & 0xC00000000)

	var ignore [2]byte
	if err = binary.Read(r, binary.LittleEndian, &ignore); err != nil {
		return
	}

	var version uint8
	if flags&int32(trackInfoVersioned) == 0 {
		version = 1
	} else {
		if err = binary.Read(r, binary.LittleEndian, &version); err != nil {
			return
		}
	}

	if err = binary.Read(r, binary.LittleEndian, &ignore); err != nil {
		return nil, err
	}

	trackInfo.TrackTitle, err = readStr(r)
	if err != nil {
		return
	}

	trackInfo.TrackAuthor, err = readStr(r)
	if err != nil {
		return
	}

	var length uint64
	if err = binary.Read(r, binary.BigEndian, &length); err != nil {
		return
	}
	trackInfo.TrackLength = time.Duration(length) * time.Millisecond

	trackInfo.TrackIdentifier, err = readStr(r)
	if err != nil {
		return nil, err
	}

	var isStream uint8
	if err = binary.Read(r, binary.LittleEndian, &isStream); err != nil {
		return
	}
	trackInfo.TrackIsStream = isStream == 1

	var hasURI uint8
	if err = binary.Read(r, binary.LittleEndian, &hasURI); err != nil {
		return nil, err
	}

	if hasURI == 1 {
		var uri string
		uri, err = readStr(r)
		trackInfo.TrackURI = &uri
		if err != nil {
			return
		}
	} else {
		trackInfo.TrackURI = nil
		_, err = readStr(r)
		if err != nil {
			return
		}
	}

	trackInfo.TrackSourceName, err = readStr(r)
	if err != nil {
		return
	}

	info = trackInfo
	return
}

func readStr(r io.Reader) (string, error) {
	var size uint16
	if err := binary.Read(r, binary.BigEndian, &size); err != nil {
		return "", err
	}
	buf := make([]byte, size)
	if err := binary.Read(r, binary.BigEndian, &buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func DefaultTracksToTracks(defaultTracks []*DefaultTrack) []Track {
	tracks := make([]Track, len(defaultTracks))
	for i := 0; i < len(defaultTracks); i++ {
		tracks[i] = defaultTracks[i]
	}
	return tracks
}
