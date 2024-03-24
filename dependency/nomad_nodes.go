// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dependency

import (
	"encoding/gob"
	"log"
	"net/url"

	"github.com/pkg/errors"
)

var (
	// Ensure NomadNodesQuery meets the Dependency interface.
	_ Dependency = (*NomadNodesQuery)(nil)
)

func init() {
	gob.Register([]*NomadNodesSnippet{})
}

// NomadNodesnippet is a stub node entry in Nomad.
type NomadNodesSnippet struct {
	ID         string
	Name       string
	Address    string
	Datacenter string
}

// NomadNodesQuery is the representation of a requested Nomad node
// dependency from inside a template.
type NomadNodesQuery struct {
	stopCh chan struct{}
}

// NewNomadNodesQuery parses a string into a NomadNodesQuery which is
// used to list nodes registered within Nomad.
func NewNomadNodesQuery(s string) (*NomadNodesQuery, error) {
	return &NomadNodesQuery{
		stopCh: make(chan struct{}, 1),
	}, nil
}

// CanShare returns a boolean if this dependency is shareable.
func (*NomadNodesQuery) CanShare() bool {
	return true
}

// Fetch queries the Nomad API defined by the given client and returns a slice
// of NomadNodesSnippet objects.
func (d *NomadNodesQuery) Fetch(clients *ClientSet, opts *QueryOptions) (interface{}, *ResponseMetadata, error) {
	select {
	case <-d.stopCh:
		return nil, nil, ErrStopped
	default:
	}

	opts = opts.Merge(&QueryOptions{})

	log.Printf("[TRACE] %s: GET %s", d, &url.URL{
		Path:     "/v1/nodes",
		RawQuery: opts.String(),
	})

	nl, qm, err := clients.Nomad().Nodes().List(opts.ToNomadOpts())
	if err != nil {
		return nil, nil, errors.Wrap(err, d.String())
	}

	log.Printf("[TRACE] %s: returned %d results", d, len(nl))

	nodes := make([]*NomadNodesSnippet, len(nl))
	for i, s := range nl {
		nodes[i] = &NomadNodesSnippet{
			Name:       s.Name,
			Datacenter: s.Datacenter,
			ID:         s.ID,
			Address:    s.Address,
		}
	}

	rm := &ResponseMetadata{
		LastIndex:   qm.LastIndex,
		LastContact: qm.LastContact,
	}

	return nodes, rm, nil
}

// String returns the human-friendly version of this dependency.
func (d *NomadNodesQuery) String() string {
	return "nomad.nodes"
}

// Stop halts the dependency's fetch function.
func (d *NomadNodesQuery) Stop() {
	close(d.stopCh)
}

// Type returns the type of this dependency.
func (d *NomadNodesQuery) Type() Type {
	return TypeNomad
}
