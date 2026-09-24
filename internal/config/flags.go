package config

import (
	"flag"
	"os"
)

// RegisterFlags adds --config and one flag per setting to fs, such as
// --storage-root for "storage.root".
func RegisterFlags(fs *flag.FlagSet) {
	fs.String("config", "", "config file (default "+DefaultConfigFile()+", or $"+ConfigEnv+")")
	for _, s := range settings {
		fs.String(FlagName(s.key), "", "set "+s.key+" (overrides $"+EnvName(s.key)+" and the config file)")
	}
}

// SourcesFromFlags returns the Sources for Load after fs, prepared with
// RegisterFlags, was parsed. Only flags given on the command line count,
// so an unset flag never hides a file or environment value.
func SourcesFromFlags(fs *flag.FlagSet) Sources {
	src := Sources{
		LookupEnv:         os.LookupEnv,
		Flags:             map[string]string{},
		DefaultConfigFile: DefaultConfigFile(),
	}
	keyByFlag := make(map[string]string, len(settings))
	for _, s := range settings {
		keyByFlag[FlagName(s.key)] = s.key
	}
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "config" {
			src.ConfigFile = f.Value.String()
			return
		}
		if key, ok := keyByFlag[f.Name]; ok {
			src.Flags[key] = f.Value.String()
		}
	})
	return src
}
