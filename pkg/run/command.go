/*
   OqtaDrive - Sinclair Microdrive emulator
   Copyright (c) 2021, Alexander Vollschwitz

   This file is part of OqtaDrive.

   OqtaDrive is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   OqtaDrive is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with OqtaDrive. If not, see <http://www.gnu.org/licenses/>.
*/

package run

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//
const (
	prologueHeader = ""
	epilogueHeader = `
Notes:

`
)

/*
	The package initializer sets up logging based on logrus. The following
	environment variables can be used to configure logging:

		LOG_FORMAT		set to `json` for JSON logging
		LOG_FORCE_COLORS	set to non-empty for forcing colorized log entries
		LOG_METHODS		set to non-empty for including methods in log
		LOG_LEVEL		`panic`, `fatal`, `error`, `warn`, `info`, `debug`, `trace`
*/
func init() {

	log.SetOutput(os.Stdout)

	if strings.ToLower(os.Getenv("LOG_FORMAT")) == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else if strings.ToLower(os.Getenv("LOG_FORCE_COLORS")) != "" {
		log.SetFormatter(&log.TextFormatter{
			ForceColors: true,
		})
	}

	if strings.ToLower(os.Getenv("LOG_METHODS")) != "" {
		log.SetReportCaller(true)
	}

	level := os.Getenv("LOG_LEVEL")
	if level != "" {
		l, err := log.ParseLevel(level)
		if err != nil {
			log.Errorf("invalid log level: '%s'; valid levels are: panic, "+
				"fatal, error, warn, info, debug, trace", level)
		} else {
			log.SetLevel(l)
		}
	}
}

//
var (
	UnderTest bool
)

// DieOnError exits the running process if e is not nil. The error gets logged.
func DieOnError(e error) {
	if e != nil {
		fmt.Printf("%v\n", e)
		if UnderTest {
			panic(e.Error())
		} else {
			os.Exit(1)
		}
	}
}

// Die exits the running process, while logging the given message.
func Die(msg string, params ...interface{}) {
	if UnderTest {
		err := fmt.Sprintf(msg, params...)
		fmt.Printf(err)
		panic(err)
	} else {
		if len(params) > 0 {
			fmt.Printf(msg, params...)
		} else {
			fmt.Println(msg)
		}
		os.Exit(1)
	}
}

//
func GetUserConfirmation(prompt string) bool {
	fmt.Printf("%s [y/N] ", prompt)
	var res string
	fmt.Scanln(&res)
	return "y" == strings.ToLower(strings.TrimSpace(res))
}

/*
	NewCommand creates a base command instance, wrapping a new Cobra command.
	The	exec function is invoked when the command's Execute method is called.
*/
func NewCommand(use, short, long, helpPrologue, helpEpilogue string,
	exec func() error) *Command {

	ret := Command{
		cmd: &cobra.Command{
			Use:   use,
			Short: short,
			Long:  long,
			RunE: func(*cobra.Command, []string) error {
				return exec()
			},
			SilenceErrors:         true,
			SilenceUsage:          true,
			DisableFlagsInUseLine: true,
		},
		settings:     map[string]*setting{},
		helpPrologue: helpPrologue,
		helpEpilogue: helpEpilogue,
	}
	ret.helpFunc = ret.cmd.HelpFunc()
	ret.cmd.SetHelpFunc(ret.help)
	return &ret
}

/*
	Command is a wrapper around Cobra & Viper. Even though they already take
	care of a lot of the boiler plate involved in getting configuration settings,
	configuration code can still get convoluted and confusing. The base command
	is here to help with this.

	Also, apparently it's not quite straightforward to have a required setting
	with Cobra/Viper that could either come from a flag or an environment
	variable (https://github.com/spf13/viper/issues/397). In addition, giving a
	meaningful error message in this case that mentions	the	flag and environment
	variable to use, is difficult.
*/
type Command struct {
	//
	cmd *cobra.Command
	//
	settings map[string]*setting
	//
	Args []string
	//
	helpPrologue string
	helpEpilogue string
	helpFunc     func(*cobra.Command, []string)
}

//
func (c *Command) help(cmd *cobra.Command, args []string) {
	if c.helpPrologue != "" {
		fmt.Fprintln(cmd.OutOrStdout(), prologueHeader+c.helpPrologue)
	}
	if c.helpFunc != nil {
		c.helpFunc(cmd, args)
	}
	if c.helpEpilogue != "" {
		fmt.Fprintln(cmd.OutOrStdout(), epilogueHeader+c.helpEpilogue)
	} else {
		fmt.Fprintln(cmd.OutOrStdout())
	}
}

/*
	Execute invokes the exec function that was set on this command when it was
	created. If args is of non-zero length, it overrides os.Args.
*/
func (c *Command) Execute(args []string) error {
	if len(args) > 0 {
		c.cmd.SetArgs(args)
	}
	return c.cmd.Execute()
}

/*
	AddSetting adds a setting to this command. Target is a pointer to the
	receiver to which the setting should be bound. Flag specifies the long
	(double-dash) command line flag for the setting, short its short
	(single-dash) version, and env the name of the environment variable that may
	carry this setting. def is a default value for the setting. When set to nil,
	the default	value will be the zero value of the setting's type. help carries
	online help	info about this setting, and required specifies whether this is
	a mandatory	setting.
*/
func (c *Command) AddSetting(target interface{}, flag, short, env string,
	def interface{}, help string, required bool) {

	s := setting{flag: flag, env: env, required: required, target: target}
	c.settings[flag] = &s

	t, n, err := s.typeAndName()
	DieOnError(err)

	log.Tracef("add setting: flag=%s, env=%s, type=%s", flag, env, t)

	if strings.HasSuffix(n, "Slice") && n != "StringSlice" && env != "" {
		Die("cannot use environment variable on non-string array setting")
	}

	if _, err := viperGetterForTypeName(n); err != nil {
		// There's a slight incongruence between pflag & Viper, that is, types
		// supported by pflag may not be supported by Viper, so we have to check
		// here to fail early on.
		Die("setting '%s' is of unsupported type: no Viper getter", flag)
	}

	defVal := reflect.Zero(t)

	if required {
		if def != nil {
			Die("required setting '%s' does not take a default value", flag)
		}
	} else if def != nil {
		if reflect.TypeOf(def).ConvertibleTo(t) {
			defVal = reflect.ValueOf(def).Convert(t)
		} else {
			Die("default value for setting '%s' has incorrect type", flag)
		}
	}

	flags := c.cmd.Flags()
	method, err := pflagMethodForTypeName(n, flags)
	if err != nil {
		Die("setting '%s' is of unsupported type: no pflag method", flag)
	}

	helpMsg := help
	if env != "" {
		helpMsg = fmt.Sprintf("%s (%s)", help, env)
	}

	method.Call(
		[]reflect.Value{
			reflect.ValueOf(target),
			reflect.ValueOf(flag),
			reflect.ValueOf(short),
			defVal,
			reflect.ValueOf(helpMsg),
		})

	viper.BindPFlag(flag, flags.Lookup(flag))
	if env != "" {
		viper.BindEnv(flag, env)
	}
}

/*
	GetSetting retrieves the setting for the provided flag and places the value
	in the variable bound to it.
*/
func (c *Command) GetSetting(flag string) (interface{}, error) {
	s, ok := c.settings[flag]
	if !ok {
		return "", fmt.Errorf("undefined setting: %s", flag)
	}
	return s.get()
}

/*
	ParseSettings handles all settings that have been added thus far via the
	AddSetting method. Afterwards, setting values are available in the variables
	to which they were bound. This should be called in the exec function that
	was	set on this command when it was created, before any references to
	variables that are bound to settings.
*/
func (c *Command) ParseSettings() {
	for _, s := range c.settings {
		_, err := s.get()
		DieOnError(err)
	}
	c.Args = c.cmd.Flags().Args()
}

//
type setting struct {
	flag     string
	env      string
	required bool
	target   interface{}
}

//
func (s *setting) typeAndName() (reflect.Type, string, error) {

	typ := reflect.TypeOf(s.target)

	if typ.Kind() != reflect.Ptr {
		return nil, "", fmt.Errorf(
			"target for setting '%s' is not a pointer", s.flag)
	}

	elem := typ.Elem()
	name := ""

	ind := reflect.Indirect(reflect.ValueOf(s.target))
	if ind.Kind() == reflect.Slice {
		name = fmt.Sprintf("%sSlice", strings.Title(ind.Type().String()[2:]))
	} else {
		name = strings.Title(elem.Name())
	}

	return elem, name, nil
}

//
func (s *setting) get() (interface{}, error) {

	t, n, err := s.typeAndName()
	if err != nil {
		return nil, err
	}

	method, err := viperGetterForTypeName(n)
	if err != nil {
		return nil, err
	}

	log.Tracef("get setting: flag=%s, type=%s", s.flag, t)
	val := method.Call([]reflect.Value{reflect.ValueOf(s.flag)})[0]
	if viper.IsSet(s.flag) {
		log.Tracef("retrieved value: '%v'", val)
	} else {
		log.Tracef("retrieved value: '%v' (default)", val)
	}

	if s.required {
		missing := false
		if val.Kind() == reflect.Slice {
			missing = val.Len() == 0
		} else {
			missing = val.Interface() == reflect.Zero(t).Interface()
		}
		if missing {
			msg := fmt.Sprintf(
				"you need to specify the --%s command line flag", s.flag)
			if s.env != "" {
				msg = fmt.Sprintf(
					"%s or the %s environment variable", msg, s.env)
			}
			return nil, fmt.Errorf("%s", msg)
		}
	}

	// Viper's BindEnv is actually not setting the variable target;
	// we need to work around this here
	if s.env != "" {
		elem := reflect.ValueOf(s.target).Elem()
		if val.Kind() == reflect.Slice {
			if elem.Len() == 0 {
				log.Trace("converting slice from env")
				elem.Set(reflect.ValueOf(stringSliceFromValue(val)))
			}
		} else {
			// We always set from val here. If the value came from a specified
			// command line flag, there will be no change. If it came from env,
			// the target still has default or zero value (due to Viper bug), so
			// we overwrite that. If neither flag nor env was set, val and
			// target have the default (if defined) or zero value set, so again
			// no change.
			elem.Set(val)
		}
	}

	return val, nil
}

//
func viperGetterForTypeName(n string) (reflect.Value, error) {
	method := fmt.Sprintf("Get%s", n)
	ret := reflect.ValueOf(viper.GetViper()).MethodByName(method)
	if ret.Kind() != reflect.Func {
		return ret, fmt.Errorf("no Viper getter %s for type %s", method, n)
	}
	return ret, nil
}

//
func pflagMethodForTypeName(n string, f *pflag.FlagSet) (reflect.Value, error) {
	method := fmt.Sprintf("%sVarP", n)
	ret := reflect.ValueOf(f).MethodByName(method)
	if ret.Kind() != reflect.Func {
		return ret, fmt.Errorf("no pflag method %s for type %s", method, n)
	}
	return ret, nil
}

//
func stringSliceFromValue(v reflect.Value) []string {
	ret := make([]string, 0, 16)
	if v.Kind() == reflect.Slice {
		for ix := 0; ix < v.Len(); ix++ {
			ret = append(ret, strings.Split(v.Index(ix).String(), ",")...)
		}
	}
	return ret
}
