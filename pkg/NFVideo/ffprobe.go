package NFVideo

import (
	"encoding/json"
	"log"
	"os/exec"
	"time"
)

type FFProbeOutput struct {
	Format  FormatDetails   `json:"format"`
	Streams []StreamDetails `json:"streams"`
}

type FormatDetails struct {
	Filename       string     `json:"filename"`
	NbStreams      int        `json:"nb_streams"`
	NbPrograms     int        `json:"nb_programs"`
	NbStreamGroups int        `json:"nb_stream_groups"`
	FormatName     string     `json:"format_name"`
	FormatLongName string     `json:"format_long_name"`
	StartTime      string     `json:"start_time"`
	Duration       string     `json:"duration"`
	Size           string     `json:"size"`
	BitRate        string     `json:"bit_rate"`
	ProbeScore     int        `json:"probe_score"`
	Tags           FormatTags `json:"tags"`
}

type FormatTags struct {
	MajorBrand       string    `json:"major_brand"`
	MinorVersion     string    `json:"minor_version"`
	CompatibleBrands string    `json:"compatible_brands"`
	CreationTime     time.Time `json:"creation_time"`
}

type StreamDetails struct {
	Index              int               `json:"index"`
	CodecName          string            `json:"codec_name"`
	CodecLongName      string            `json:"codec_long_name"`
	Profile            string            `json:"profile"`
	CodecType          string            `json:"codec_type"`
	CodecTagString     string            `json:"codec_tag_string"`
	CodecTag           string            `json:"codec_tag"`
	Width              int               `json:"width"`
	Height             int               `json:"height"`
	CodedWidth         int               `json:"coded_width"`
	CodedHeight        int               `json:"coded_height"`
	ClosedCaptions     int               `json:"closed_captions"`
	FilmGrain          int               `json:"film_grain"`
	HasBFrames         int               `json:"has_b_frames"`
	SampleAspectRatio  string            `json:"sample_aspect_ratio"`
	DisplayAspectRatio string            `json:"display_aspect_ratio"`
	PixFmt             string            `json:"pix_fmt"`
	Level              int               `json:"level"`
	ColorRange         string            `json:"color_range"`
	ColorSpace         string            `json:"color_space"`
	ColorTransfer      string            `json:"color_transfer"`
	ColorPrimaries     string            `json:"color_primaries"`
	ChromaLocation     string            `json:"chroma_location"`
	FieldOrder         string            `json:"field_order"`
	Refs               int               `json:"refs"`
	IsAvc              string            `json:"is_avc"`
	NalLengthSize      string            `json:"nal_length_size"`
	Id                 string            `json:"id"`
	RFrameRate         string            `json:"r_frame_rate"`
	AvgFrameRate       string            `json:"avg_frame_rate"`
	TimeBase           string            `json:"time_base"`
	StartPts           int               `json:"start_pts"`
	StartTime          string            `json:"start_time"`
	DurationTs         int               `json:"duration_ts"`
	Duration           string            `json:"duration"`
	BitRate            string            `json:"bit_rate"`
	BitsPerRawSample   string            `json:"bits_per_raw_sample"`
	NbFrames           string            `json:"nb_frames"`
	ExtradataSize      int               `json:"extradata_size"`
	Disposition        StreamDisposition `json:"disposition"`
	Tags               StreamTags        `json:"tags"`
}

type StreamDisposition struct {
	Default         int `json:"default"`
	Dub             int `json:"dub"`
	Original        int `json:"original"`
	Comment         int `json:"comment"`
	Lyrics          int `json:"lyrics"`
	Karaoke         int `json:"karaoke"`
	Forced          int `json:"forced"`
	HearingImpaired int `json:"hearing_impaired"`
	VisualImpaired  int `json:"visual_impaired"`
	CleanEffects    int `json:"clean_effects"`
	AttachedPic     int `json:"attached_pic"`
	TimedThumbnails int `json:"timed_thumbnails"`
	NonDiegetic     int `json:"non_diegetic"`
	Captions        int `json:"captions"`
	Descriptions    int `json:"descriptions"`
	Metadata        int `json:"metadata"`
	Dependent       int `json:"dependent"`
	StillImage      int `json:"still_image"`
}

type StreamTags struct {
	CreationTime time.Time `json:"creation_time"`
	Language     string    `json:"language"`
	HandlerName  string    `json:"handler_name"`
	VendorId     string    `json:"vendor_id"`
}

func ProbeVideo(file string, args ...string) (FFProbeOutput, error) {
	cmd := exec.Command(probePath, file)
	for _, arg := range args {
		if arg != "" {
			cmd.Args = append(cmd.Args, arg)
		}
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Could not run ffprobe")
		return FFProbeOutput{}, err
	}

	var probeOutput FFProbeOutput
	if err = json.Unmarshal(output, &probeOutput); err != nil {
		log.Println("Could not unmarshal the ffprobe output")
		return FFProbeOutput{}, err
	}

	return probeOutput, nil
}
