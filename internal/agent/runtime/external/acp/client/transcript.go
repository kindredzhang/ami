package client

import "github.com/kindredzhang/ami/internal/agent/runtime/external"

type TranscriptRecorder = external.TranscriptRecorder

var (
	NewTranscriptRecorder = external.NewTranscriptRecorder
	TranscriptFromEvents  = external.TranscriptFromEvents
)
