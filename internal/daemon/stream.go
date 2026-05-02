package daemon

import api "github.com/raphael-goetz/lazysound/lib/soundcloud"

func chooseStreamURL(s *api.TrackStreams) string {
	if s == nil {
		return ""
	}
	if s.HLSAAC160URL != "" {
		return s.HLSAAC160URL
	}
	if s.HLSAAC96URL != "" {
		return s.HLSAAC96URL
	}
	if s.HTTPMP3128URL != "" {
		return s.HTTPMP3128URL
	}
	if s.HLSMP3128URL != "" {
		return s.HLSMP3128URL
	}
	if s.HLSOPUS64URL != "" {
		return s.HLSOPUS64URL
	}
	if s.PreviewMP3128 != "" {
		return s.PreviewMP3128
	}
	return ""
}
