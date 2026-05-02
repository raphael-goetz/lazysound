package daemon

import api "github.com/raphael-goetz/lazysound/lib/soundcloud"

type streamCandidate struct {
	kind string
	url  string
}

func streamCandidates(s *api.TrackStreams) []streamCandidate {
	if s == nil {
		return nil
	}
	var out []streamCandidate
	if s.HLSAAC160URL != "" {
		out = append(out, streamCandidate{kind: "hls_aac_160", url: s.HLSAAC160URL})
	}
	if s.HLSAAC96URL != "" {
		out = append(out, streamCandidate{kind: "hls_aac_96", url: s.HLSAAC96URL})
	}
	if s.HTTPMP3128URL != "" {
		out = append(out, streamCandidate{kind: "http_mp3_128", url: s.HTTPMP3128URL})
	}
	if s.HLSMP3128URL != "" {
		out = append(out, streamCandidate{kind: "hls_mp3_128", url: s.HLSMP3128URL})
	}
	if s.HLSOPUS64URL != "" {
		out = append(out, streamCandidate{kind: "hls_opus_64", url: s.HLSOPUS64URL})
	}
	return out
}
