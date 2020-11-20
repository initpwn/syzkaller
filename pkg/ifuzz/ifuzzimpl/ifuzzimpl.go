// Copyright 2017 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package ifuzzimpl

import (
	"math/rand"
)

const (
	ArchX86     = "x86"
	ArchPowerPC = "powerpc"
)

var Arches = make(map[string]InsnSet)

type (
	Mode int
	Type int
)

type Insn interface {
	GetName() string
	GetMode() int
	GetPseudo() bool
	GetPriv() bool
	IsCompatible(cfg *Config) bool
	Encode(cfg *Config, r *rand.Rand) []byte
}

type InsnSet interface {
	GetInsns(mode Mode, typ Type) []Insn
	Decode(mode Mode, text []byte) (int, error)
	DecodeExt(mode Mode, text []byte) (int, error) // XED, to keep ifuzz_test happy
}

type Config struct {
	Arch       string
	Len        int         // number of instructions to generate
	Mode       Mode        // one of ModeXXX
	Priv       bool        // generate CPL=0 instructions (x86), HV/!PR mode (PPC)
	Exec       bool        // generate instructions sequences interesting for execution
	MemRegions []MemRegion // generated instructions will reference these regions
}

type MemRegion struct {
	Start uint64
	Size  uint64
}

const (
	ModeLong64 Mode = iota
	ModeProt32
	ModeProt16
	ModeReal16
	ModeLast
)

const (
	TypeExec Type = iota
	TypePriv
	TypeUser
	TypeAll
	TypeLast
)

// ModeInsns returns list of all instructions for the given mode.
func ModeInsns(cfg *Config) []Insn {
	insnset := Arches[cfg.Arch]
	if cfg.Mode < 0 || cfg.Mode >= ModeLast {
		panic("bad mode")
	}
	insns := insnset.GetInsns(cfg.Mode, TypeUser)
	if cfg.Priv {
		insns = append(insns, insnset.GetInsns(cfg.Mode, TypePriv)...)
		if cfg.Exec {
			insns = append(insns, insnset.GetInsns(cfg.Mode, TypeExec)...)
		}
	}
	return insns
}

var SpecialNumbers = [...]uint64{0, 1 << 15, 1 << 16, 1 << 31, 1 << 32, 1 << 47, 1 << 47, 1 << 63}