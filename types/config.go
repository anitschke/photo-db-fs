package types

import (
	"errors"
	"fmt"
	"strings"
)

// xxx doc
type Config struct {
	MountPoint string        `json:"mountPoint,omitempty"`
	DB         DB            `json:"db"`
	LogLevel   string        `json:"logLevel,omitempty"`
	Queries    []QueryConfig `json:"queries,omitempty"`
}

// xxx doc
type DB struct {
	Type   string `json:"type,omitempty"`
	Source string `json:"source,omitempty"`
}

// xxx doc
type QueryConfig struct {
	Name     string         `json:"name,omitempty"`
	Selector SelectorConfig `json:"selector"`
}

type SelectorPropertyMap map[string]SelectorProperty

// xxx doc
type SelectorConfig struct {
	Type       string              `json:"type,omitempty"`
	Properties SelectorPropertyMap `json:"properties,omitempty"`
}

// xxx doc
type SelectorProperty struct {
	Strings   []string         `json:"strings,omitempty"`
	Selectors []SelectorConfig `json:"selectors,omitempty"`
	Selector  *SelectorConfig  `json:"selector,omitempty"`
}

func ConfigToQuery(config QueryConfig) (NamedQuery, error) {
	s, err := configToSelector(config.Selector)
	if err != nil {
		return NamedQuery{}, fmt.Errorf("error parsing config %q: %w", config.Name, err)
	}
	return NamedQuery{
		Name: config.Name,
		Query: Query{
			Selector: s,
		},
	}, nil
}

func ConfigsToQueries(configs []QueryConfig) ([]NamedQuery, error) {
	namedQueries := make([]NamedQuery, 0, len(configs))
	for _, c := range configs {
		q, err := ConfigToQuery(c)
		if err != nil {
			return nil, err
		}
		namedQueries = append(namedQueries, q)
	}
	return namedQueries, nil
}

func ValidateQueryConfigs(configs []QueryConfig) error {
	names := make(map[string]struct{}, len(configs))
	for _, c := range configs {
		_, ok := names[c.Name]
		if ok {
			return fmt.Errorf("query configs must have unique names, the name %q is used for more than one query", c.Name)
		}
		// TODO validate that c.Name is a valid file name
		// https://github.com/anitschke/photo-db-fs/issues/1
	}
	return nil
}

func configToSelector(config SelectorConfig) (Selector, error) {
	switch t := strings.ToLower(config.Type); t {
	case "hastag": // cspell:disable-line
		return configToHasTag(config)
	case "and":
		return configToAnd(config)
	case "or":
		return configToOr(config)
	case "difference":
		return configToDifference(config)
	default:
		return nil, fmt.Errorf("invalid selector type %q", config.Type)
	}
}

func configToHasTag(config SelectorConfig) (Selector, error) {
	var s HasTag
	for name, p := range config.Properties {
		switch n := strings.ToLower(name); n {
		case "tag":
			s.Tag.Path = p.Strings
		default:
			return nil, fmt.Errorf("invalid property %q", name)
		}
	}
	return s, nil
}

func configToAnd(config SelectorConfig) (Selector, error) {
	var s And
	for name, p := range config.Properties {
		switch n := strings.ToLower(name); n {
		case "operands":
			s.Operands = make([]Selector, 0, len(p.Selectors))
			for _, opConfig := range p.Selectors {
				op, err := configToSelector(opConfig)
				if err != nil {
					return nil, err
				}
				s.Operands = append(s.Operands, op)
			}
		default:
			return nil, fmt.Errorf("invalid property %q", name)
		}
	}
	return s, nil
}

func configToOr(config SelectorConfig) (Selector, error) {
	var s Or
	for name, p := range config.Properties {
		switch n := strings.ToLower(name); n {
		case "operands":
			s.Operands = make([]Selector, 0, len(p.Selectors))
			for _, opConfig := range p.Selectors {
				op, err := configToSelector(opConfig)
				if err != nil {
					return nil, err
				}
				s.Operands = append(s.Operands, op)
			}
		default:
			return nil, fmt.Errorf("invalid property %q", name)
		}
	}
	return s, nil
}

func configToDifference(config SelectorConfig) (Selector, error) {
	var s Difference
	for name, p := range config.Properties {
		switch n := strings.ToLower(name); n {
		case "starting":
			if p.Selector == nil {
				return nil, errors.New("unspecified starting selector")
			}
			start, err := configToSelector(*p.Selector)
			if err != nil {
				return nil, err
			}
			s.Starting = start
		case "excluding":
			if p.Selector == nil {
				return nil, errors.New("unspecified excluding selector")
			}
			exclude, err := configToSelector(*p.Selector)
			if err != nil {
				return nil, err
			}
			s.Excluding = exclude
		default:
			return nil, fmt.Errorf("invalid property %q", name)
		}
	}
	return s, nil
}
