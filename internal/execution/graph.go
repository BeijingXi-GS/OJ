// Package execution declares a draft process graph; no execution is implemented.
package execution

import "oj.local/oj/internal/domain"

type SourceBinding struct {
	SourceUnitID domain.ID
	ArtifactID   domain.ID
}

type ProcessSpec struct {
	ID      string
	Role    string
	Binding SourceBinding
}

type ChannelSpec struct {
	ID            string
	FromProcessID string
	ToProcessID   string
	ByteLimit     int64
	MessageLimit  int64
}

type Graph struct {
	Processes        []ProcessSpec
	Channels         []ChannelSpec
	Resources        domain.ResourcePolicy
	EvaluatorVersion string
}
