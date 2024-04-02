package dependency

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"regexp"

	"github.com/pkg/errors"
)

var (
	// Ensure NomadNodeQuery meets the Dependency interface.
	_ Dependency = (*NomadNodeQuery)(nil)

	// NomadNodeQueryRe is the regular expression for querying a Nomad node.
	NomadNodeQueryRe = regexp.MustCompile(`\A` + nodeNameRe + `\z`)
)

// NomadNodeQuery is the representation of a requested Nomad node
// dependency from inside a template.
type NomadNodeQuery struct {
	stopCh chan struct{}
	id     string
}

// NewNomadNodeQuery parses a string into a NomadNodesQuery which is
// used to list nodes registered within Nomad.
func NewNomadNodeQuery(s string) (*NomadNodeQuery, error) {
	if s != "" && !NomadNodeQueryRe.MatchString(s) {
		return nil, fmt.Errorf("nomad.nodes: invalid format: %q", s)
	}

	m := regexpMatch(NomadNodeQueryRe, s)

	return &NomadNodeQuery{
		stopCh: make(chan struct{}, 1),
		id:     m["name"],
	}, nil
}

// CanShare returns a boolean if this dependency is shareable.
func (*NomadNodeQuery) CanShare() bool {
	return true
}

// Fetch queries the Nomad API defined by the given client and returns a slice
// of NomadNodesSnippet objects.
func (d *NomadNodeQuery) Fetch(clients *ClientSet, opts *QueryOptions) (interface{}, *ResponseMetadata, error) {
	select {
	case <-d.stopCh:
		return nil, nil, ErrStopped
	default:
	}

	log.Printf("[TRACE] %s: GET %s", d, &url.URL{
		Path: "/v1/node/" + d.id,
	})

	n, qm, err := clients.Nomad().Nodes().Info(d.id, opts.ToNomadOpts())
	if err != nil {
		return nil, nil, errors.Wrap(err, d.String())
	}

	log.Printf("[TRACE] %s: returned node info", d)

	addr := n.Attributes["nomad.advertise.address"]
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, nil, errors.Wrap(err, d.String())
	}

	ns := &NomadNodeSnippet{
		Name:       n.Name,
		Datacenter: n.Datacenter,
		ID:         n.ID,
		Address:    host,
	}

	rm := &ResponseMetadata{
		LastIndex:   qm.LastIndex,
		LastContact: qm.LastContact,
	}

	return ns, rm, nil
}

// String returns the human-friendly version of this dependency.
func (d *NomadNodeQuery) String() string {
	if d.id != "" {
		return fmt.Sprintf("nomad.node(@%s)", d.id)
	}
	return "nomad.node"
}

// Stop halts the dependency's fetch function.
func (d *NomadNodeQuery) Stop() {
	close(d.stopCh)
}

// Type returns the type of this dependency.
func (d *NomadNodeQuery) Type() Type {
	return TypeNomad
}
