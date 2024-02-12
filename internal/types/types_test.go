//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package types_test

import (
	"fmt"
	"testing"

	"github.com/zalando/rds-health/internal/types"
)

func TestBiB(t *testing.T) {
	for value, expected := range map[types.BiB]string{
		types.BiB(1):  "1 bytes",
		types.BiB(5):  "5 bytes",
		types.KiB:     "1 KiB",
		5 * types.KiB: "5 KiB",
		types.MiB:     "1 MiB",
		5 * types.MiB: "5 MiB",
		types.GiB:     "1 GiB",
		5 * types.GiB: "5 GiB",
		types.TiB:     "1 TiB",
		5 * types.TiB: "5 TiB",
	} {
		check(t, value, expected)
	}
}

func TestGHz(t *testing.T) {
	for value, expected := range map[types.GHz]string{
		1.0:   "1.00 GHz",
		1.1:   "1.10 GHz",
		1.123: "1.12 GHz",
		1.127: "1.13 GHz",
	} {
		check(t, value, expected)
	}
}

func TestStorage(t *testing.T) {
	for value, expected := range map[types.Storage]string{
		{"memory", 4 * types.GiB}: "mem 4 GiB",
	} {
		check(t, value, expected)
	}
}

func TestCPU(t *testing.T) {
	for value, expected := range map[types.CPU]string{
		{4, types.GHz(2.2)}: "4 vcpu 2.20 GHz",
	} {
		check(t, value, expected)
	}
}

func TestCompute(t *testing.T) {
	for value, expected := range map[types.Compute]string{
		{}: "",
		{
			CPU: &types.CPU{4, 1.2},
		}: "4 vcpu 1.20 GHz",
		{
			CPU:    &types.CPU{4, 1.2},
			Memory: &types.Storage{"memory", 16 * types.GiB},
		}: "4 vcpu 1.20 GHz, mem 16 GiB",
	} {
		check(t, value, expected)
	}
}

func TestEngine(t *testing.T) {
	for value, expected := range map[types.Engine]string{
		{"aurora", "3.4.5"}: "aurora v3.4.5",
	} {
		check(t, value, expected)
	}
}

func TestNode(t *testing.T) {
	for value, expected := range map[*types.Node]string{
		{
			ID: "x",
		}: "",
		{
			Type: "db.m5d.large",
		}: "db.m5d.large",
		{
			Type:   "db.m5d.large",
			Engine: &types.Engine{"aurora", "3.4.5"},
		}: "db.m5d.large aurora v3.4.5",
		{
			Type:    "db.m5d.large",
			Engine:  &types.Engine{"aurora", "3.4.5"},
			Compute: &types.Compute{CPU: &types.CPU{4, 1.2}, Memory: &types.Storage{"memory", 16 * types.GiB}},
		}: "db.m5d.large aurora v3.4.5 (4 vcpu 1.20 GHz, mem 16 GiB)",
		{
			Type:    "db.m5d.large",
			Engine:  &types.Engine{"aurora", "3.4.5"},
			Compute: &types.Compute{CPU: &types.CPU{4, 1.2}, Memory: &types.Storage{"memory", 16 * types.GiB}},
			Storage: &types.Storage{"io1", 100 * types.GiB},
		}: "db.m5d.large aurora v3.4.5 (4 vcpu 1.20 GHz, mem 16 GiB, storage io1 100 GiB)",
	} {
		check(t, value, expected)
	}

}

//
// Helper
//

func check[T interface{ String() string }](t *testing.T, value T, expected string) {
	t.Helper()

	s := fmt.Sprintf("%s", value)
	if s != expected {
		t.Errorf("%s != %s", value, expected)
	}
}
