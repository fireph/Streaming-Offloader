package main

import (
    "context"
    "log"
    "os"
    "os/exec"
    "time"
    "gopkg.in/yaml.v2"
)

type StreamConfig struct {
    Name       string `yaml:"name"`
    URL        string `yaml:"url"`
    Key        string `yaml:"key"`
    Video      CodecConfig `yaml:"video"`
    Audio      CodecConfig `yaml:"audio"`
    Passthrough bool   `yaml:"passthrough"`
}

type CodecConfig struct {
    Codec   string `yaml:"codec"`
    Bitrate string `yaml:"bitrate"`
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
    dec := yaml.NewDecoder(f)
    err = dec.Decode(&cfg)
    return &cfg, err
}

func main() {
    cfg, err := loadConfig("/config/config.yaml")
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }
    // build ffmpeg args
    baseArgs := []string{
        "-f", "flv", "-listen", "1",
        "-timeout", fmt.Sprint(cfg.TimeoutSec * 1000000),
        fmt.Sprintf("tcp://0.0.0.0:%d", cfg.ListenPort),
    }
    for _, s := range cfg.Streams {
        out := fmt.Sprintf("%s/%s", s.URL, s.Key)
        args := append([]string{}, baseArgs...)
        if s.Passthrough {
            args = append(args, "-c", "copy")
        } else {
            args = append(args,
                "-c:v", s.Video.Codec,
                "-b:v", s.Video.Bitrate,
                "-c:a", s.Audio.Codec,
                "-b:a", s.Audio.Bitrate,
            )
        }
        args = append(args, "-f", "flv", out)
        go func(name string) {
            for {
                ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSec+10)*time.Second)
                defer cancel()
                cmd := exec.CommandContext(ctx, "ffmpeg", args...)
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
                log.Printf("starting stream %s", name)
                err := cmd.Run()
                if err != nil {
                    log.Printf("stream %s exited: %v", name, err)
                }
                // restart after short delay
                time.Sleep(5 * time.Second)
            }
        }(s.Name)
    }
    // block forever
    select {}
}
