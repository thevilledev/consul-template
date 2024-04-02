// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dependency

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewNomadNodesQuery(t *testing.T) {
	cases := []struct {
		name string
		i    string
		exp  *NomadNodesQuery
		err  bool
	}{
		{
			"empty",
			"",
			&NomadNodesQuery{},
			false,
		},
		{
			"string",
			"string",
			nil,
			true,
		},
		{
			"datacenter",
			"@dc1",
			&NomadNodesQuery{
				datacenter: "dc1",
			},
			false,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			act, err := NewNomadNodesQuery(tc.i)
			if (err != nil) != tc.err {
				t.Fatal(err)
			}

			if act != nil {
				act.stopCh = nil
			}

			require.Equal(t, tc.exp, act)
		})
	}
}

func TestNomadNodesQuery_String(t *testing.T) {
	cases := []struct {
		name string
		i    string
		exp  string
	}{
		{
			"empty",
			"",
			"nomad.nodes",
		},
		{
			"datacenter",
			"@us-east-1a",
			"nomad.nodes(@us-east-1a)",
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			d, err := NewNomadNodesQuery(tc.i)
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, tc.exp, d.String())
		})
	}
}
