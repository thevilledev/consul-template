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
	NomadNodesQueryRe = regexp.MustCompile(`\A` + regionRe + `\z`)
)

func init() {
	gob.Register([]*NomadNodesSnippet{})
}

// NomadNodesSnippet is a stub node entry in Nomad.
type NomadNodesSnippet struct {
	ID         string
	Name       string
	Address    string
	Datacenter string
	Region     string
}

// NomadNodesQuery is the representation of a requested Nomad node
// dependency from inside a template.
type NomadNodesQuery struct {
	stopCh chan struct{}
	region string
}

// NewNomadNodesQuery parses a string into a NomadNodesQuery which is
// used to list nodes registered within Nomad.
func NewNomadNodesQuery(s string) (*NomadNodesQuery, error) {
	if s != "" && !NomadNodesQueryRe.MatchString(s) {
		return nil, fmt.Errorf("nomad.nodes: invalid format: %q", s)
	}

	m := regexpMatch(NomadNodesQueryRe, s)

	return &NomadNodesQuery{
		stopCh: make(chan struct{}, 1),
		region: m["region"],
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
		Region: d.region,
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

	nodes := make([]*NomadNodesSnippet, len(nl))
	for i, s := range nl {
		nodes[i] = &NomadNodesSnippet{
			Name:       s.Name,
			Datacenter: s.Datacenter,
			ID:         s.ID,
			Address:    s.Address,
			Region:     d.region,
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
	if d.region != "" {
		return fmt.Sprintf("nomad.nodes(@%s)", d.region)
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
type NomadNodesByName []*NomadNodesSnippet

func (s NomadNodesByName) Len() int           { return len(s) }
func (s NomadNodesByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s NomadNodesByName) Less(i, j int) bool { return s[i].Name < s[j].Name }
