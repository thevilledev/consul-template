// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dependency

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"sort"

	"github.com/pkg/errors"
)

var (
	// Ensure NomadNodesQuery meets the Dependency interface.
	_ Dependency = (*NomadNodesQuery)(nil)

	// NomadNodesQueryRe is the regular expression for querying Nomad nodes.
	NomadNodesQueryRe = regexp.MustCompile(`\A` + dcRe + `\z`)
)

func init() {
	gob.Register([]*NomadNodeSnippet{})
}

// NomadNodeSnippet is a stub node entry in Nomad.
type NomadNodeSnippet struct {
	ID         string
	Name       string
	Address    string
	Datacenter string
}

// NomadNodesQuery is the representation of a requested Nomad node
// dependency from inside a template.
type NomadNodesQuery struct {
	stopCh     chan struct{}
	datacenter string
}

// NewNomadNodesQuery parses a string into a NomadNodesQuery which is
// used to list nodes registered within Nomad.
func NewNomadNodesQuery(s string) (*NomadNodesQuery, error) {
	if s != "" && !NomadNodesQueryRe.MatchString(s) {
		return nil, fmt.Errorf("nomad.nodes: invalid format: %q", s)
	}

	m := regexpMatch(NomadNodesQueryRe, s)

	return &NomadNodesQuery{
		stopCh:     make(chan struct{}, 1),
		datacenter: m["dc"],
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

	opts = opts.Merge(&QueryOptions{
		Filter: fmt.Sprintf("Datacenter == %s", d.datacenter),
	})

	log.Printf("[TRACE] %s: GET %s", d, &url.URL{
		Path:     "/v1/nodes",
		RawQuery: opts.String(),
	})

	nl, qm, err := clients.Nomad().Nodes().List(opts.ToNomadOpts())
	if err != nil {
		return nil, nil, errors.Wrap(err, d.String())
	}

	log.Printf("[TRACE] %s: returned %d results", d, len(nl))

	nodes := make([]*NomadNodeSnippet, len(nl))
	for i, s := range nl {
		nodes[i] = &NomadNodeSnippet{
			Name:       s.Name,
			Datacenter: s.Datacenter,
			ID:         s.ID,
			Address:    s.Address,
		}
	}

	sort.Stable(NomadNodesByName(nodes))

	rm := &ResponseMetadata{
		LastIndex:   qm.LastIndex,
		LastContact: qm.LastContact,
	}

	return nodes, rm, nil
}

// String returns the human-friendly version of this dependency.
func (d *NomadNodesQuery) String() string {
	if d.datacenter != "" {
		return fmt.Sprintf("nomad.nodes(@%s)", d.datacenter)
	}
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

// NomadNodesByName is a sortable slice of CatalogService structs.
type NomadNodesByName []*NomadNodeSnippet

func (s NomadNodesByName) Len() int           { return len(s) }
func (s NomadNodesByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s NomadNodesByName) Less(i, j int) bool { return s[i].Name < s[j].Name }
