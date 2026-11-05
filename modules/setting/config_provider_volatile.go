// Copyright 2023 The Gitea Authors. All rights reserved.
// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT
//
// This extension of the ConfigProvider interface is to test saving config
// changes from unit tests without involving file operations.
//
// it simply wraps ConfigProvider and NOOPs the save methods
//
// if the wrapped Provider is an iniConfigProvider, it implements WriteTo and,
// based on that, String() as a handy tool

package setting

import (
	"bytes"
	"io"
)

type VolatileConfigProvider interface {
	ConfigProvider
	WriteTo(w io.Writer) (int64, error)
	String() (string, error)
}

type volatileConfig struct {
	config ConfigProvider
}

func NewVolatileConfigProvider(config ConfigProvider, err error) (VolatileConfigProvider, error) {
	if err != nil {
		return nil, err
	}
	return &volatileConfig{config}, nil
}

// added methods
func (v *volatileConfig) WriteTo(w io.Writer) (int64, error) {
	return v.config.(*iniConfigProvider).ini.WriteTo(w)
}

func (v *volatileConfig) String() (string, error) {
	var buf bytes.Buffer
	_, err := v.WriteTo(&buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// noop methods
func (v *volatileConfig) Save() error {
	return nil
}

func (v *volatileConfig) SaveTo(filename string) error {
	return nil
}

func (v *volatileConfig) GetFile() string {
	return ""
}

func (v *volatileConfig) DisableSaving() {
}

func (v *volatileConfig) PrepareSaving() (ConfigProvider, error) {
	return v, nil
}

// 1:1 wrappers
func (v *volatileConfig) Section(section string) ConfigSection {
	return v.config.Section(section)
}

func (v *volatileConfig) Sections() []ConfigSection {
	return v.config.Sections()
}

func (v *volatileConfig) NewSection(name string) (ConfigSection, error) {
	return v.config.NewSection(name)
}

func (v *volatileConfig) GetSection(name string) (ConfigSection, error) {
	return v.config.GetSection(name)
}
