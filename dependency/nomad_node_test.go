// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dependency

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewNomadNodeQuery(t *testing.T) {
	cases := []struct {
		name string
		i    string
		exp  *NomadNodeQuery
		err  bool
	}{
		{
			"empty",
			"",
			nil,
			true,
		},
		{
			"string",
			"string",
			&NomadNodeQuery{
				id: "string",
			},
			false,
		},
		{
			"datacenter",
			"@dc1",
			nil,
			true,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			act, err := NewNomadNodeQuery(tc.i)
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

func TestNomadNodeQuery_String(t *testing.T) {
	cases := []struct {
		name string
		i    string
		exp  string
	}{
		{
			"id",
			"foo",
			"nomad.node(foo)",
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			d, err := NewNomadNodeQuery(tc.i)
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, tc.exp, d.String())
		})
	}
}
