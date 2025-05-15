package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/exec"
    "strconv"
    "time"

    "gopkg.in/yaml.v2"
)

type CodecConfig struct {
    Codec          string `yaml:"codec"`
    Bitrate        string `yaml:"bitrate"`
    PixelFormat    string `yaml:"pixel_format,omitempty"`
    RateControl    string `yaml:"rate_control,omitempty"`
    Preset         string `yaml:"preset,omitempty"`
    KeyInt         int    `yaml:"keyint,omitempty"`
    Tune           string `yaml:"tune,omitempty"`
    Profile        string `yaml:"profile,omitempty"`
    LookaheadLevel int    `yaml:"lookahead_level,omitempty"`
    SpatialAQ      bool   `yaml:"spatial_aq,omitempty"`
    TemporalAQ     bool   `yaml:"temporal_aq,omitempty"`
    BFrames        int    `yaml:"bframes,omitempty"`
    BRefMode       string `yaml:"b_ref_mode,omitempty"`
    Multipass      string `yaml:"multipass,omitempty"`
}

type StreamConfig struct {
    Name  string      `yaml:"name"`
    URL   string      `yaml:"url"`
    Key   string      `yaml:"key"`
    Video CodecConfig `yaml:"video"`
    Audio CodecConfig `yaml:"audio"`
}

type Config struct {
    ListenPort int            `yaml:"listen_port"`
    TimeoutSec int            `yaml:"timeout_sec"`
    Streams    []StreamConfig `yaml:"streams"`
}

func loadConfig(path string) (*Config, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    var cfg Config
    if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}

func buildArgs(s StreamConfig) []string {
    args := []string{"-f", "flv", "-listen", "1"}
    // Video: copy or encode with NVENC options
    if s.Video.Codec == "copy" {
        args = append(args, "-c:v", "copy")
    } else {
        args = append(args, "-c:v", s.Video.Codec, "-b:v", s.Video.Bitrate)
        if s.Video.PixelFormat != "" {
            args = append(args, "-pix_fmt", s.Video.PixelFormat)
        }
        if s.Video.RateControl != "" {
            args = append(args, "-rc:v", s.Video.RateControl)
        }
        if s.Video.Preset != "" {
            args = append(args, "-preset", s.Video.Preset)
        }
        if s.Video.Tune != "" {
            args = append(args, "-tune", s.Video.Tune)
        }
        if s.Video.Profile != "" {
            args = append(args, "-profile:v", s.Video.Profile)
        }
        if s.Video.KeyInt > 0 {
            args = append(args, "-g", strconv.Itoa(s.Video.KeyInt))
        }
        if s.Video.LookaheadLevel > 0 {
            args = append(args, "-lookahead", strconv.Itoa(s.Video.LookaheadLevel))
        }
        if s.Video.SpatialAQ {
            args = append(args, "-spatial-aq", "1")
        }
        if s.Video.TemporalAQ {
            args = append(args, "-temporal-aq", "1")
        }
        if s.Video.BFrames > 0 {
            args = append(args, "-bf", strconv.Itoa(s.Video.BFrames))
        }
        if s.Video.BRefMode != "" {
            args = append(args, "-b_ref_mode", s.Video.BRefMode)
        }
        if s.Video.Multipass != "" {
            args = append(args, "-multipass", s.Video.Multipass)
        }
    }
    // Audio: copy or basic encode
    if s.Audio.Codec == "copy" {
        args = append(args, "-c:a", "copy")
    } else {
        args = append(args, "-c:a", s.Audio.Codec, "-b:a", s.Audio.Bitrate)
    }
    args = append(args, "-f", "flv", fmt.Sprintf("%s/%s", s.URL, s.Key))
    return args
}

func main() {
    cfg, err := loadConfig("/config/config.yaml")
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }
    for _, s := range cfg.Streams {
        args := buildArgs(s)
        go func(name string, cmdArgs []string) {
            for {
                ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSec+10)*time.Second)
                cmd := exec.CommandContext(ctx, "ffmpeg", cmdArgs...)
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
                log.Printf("starting stream %s", name)
                err := cmd.Run()
                cancel()
                if err != nil {
                    log.Printf("stream %s exited: %v", name, err)
                }
                time.Sleep(5 * time.Second)
            }
        }(s.Name, args)
    }
    select {}
}
