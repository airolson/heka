/***** BEGIN LICENSE BLOCK *****
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this file,
# You can obtain one at http://mozilla.org/MPL/2.0/.
#
# The Initial Developer of the Original Code is the Mozilla Foundation.
# Portions created by the Initial Developer are Copyright (C) 2012
# the Initial Developer. All Rights Reserved.
#
# Contributor(s):
#   Rob Miller (rmiller@mozilla.com)
#
# ***** END LICENSE BLOCK *****/

package pipeline

import (
	"github.com/mozilla-services/heka/message"
	ts "github.com/mozilla-services/heka/testsupport"
	gs "github.com/rafrombrc/gospec/src/gospec"
)

func LoadFromConfigSpec(c gs.Context) {
	c.Specify("Config file loading", func() {
		origGlobals := Globals
		origDecodersByEncoding := DecodersByEncoding
		pipeConfig := NewPipelineConfig(nil)
		defer func() {
			Globals = origGlobals
			DecodersByEncoding = origDecodersByEncoding
		}()

		c.Assume(pipeConfig, gs.Not(gs.IsNil))

		c.Specify("works w/ good config file", func() {
			err := pipeConfig.LoadFromConfigFile("../testsupport/config_test.toml")
			c.Assume(err, gs.IsNil)

			// We use a set of Expect's rather than c.Specify because the
			// pipeConfig can't be re-loaded per child as gospec will do
			// since each one needs to bind to the same address

			// and the decoders are loaded for the right encoding headers
			dSet := pipeConfig.DecoderSets[0]
			dRunner, ok := dSet.ByEncoding(message.Header_JSON)
			c.Expect(dRunner, gs.Not(gs.IsNil))
			c.Expect(ok, gs.IsTrue)
			c.Expect(dRunner.Name(), gs.Equals, "JsonDecoder")

			dRunner, ok = dSet.ByEncoding(message.Header_PROTOCOL_BUFFER)
			c.Expect(dRunner, gs.Not(gs.IsNil))
			c.Expect(ok, gs.IsTrue)
			c.Expect(dRunner.Name(), gs.Equals, "ProtobufDecoder")

			// decoders channel is full
			c.Expect(len(pipeConfig.decodersChan), gs.Equals, Globals().DecoderPoolSize)

			// and the inputs section loads properly with a custom name
			_, ok = pipeConfig.InputRunners["UdpInput"]
			c.Expect(ok, gs.Equals, true)

			// and the decoders sections load
			_, ok = pipeConfig.DecoderWrappers["JsonDecoder"]
			c.Expect(ok, gs.Equals, true)
			_, ok = pipeConfig.DecoderWrappers["ProtobufDecoder"]
			c.Expect(ok, gs.Equals, true)

			// and the outputs section loads
			_, ok = pipeConfig.OutputRunners["LogOutput"]
			c.Expect(ok, gs.Equals, true)

			// and the filters sections loads
			_, ok = pipeConfig.FilterRunners["sample"]
			c.Expect(ok, gs.Equals, true)
		})

		c.Specify("works w/ decoder defaults", func() {
			err := pipeConfig.LoadFromConfigFile("../testsupport/config_test_defaults.toml")
			c.Assume(err, gs.Not(gs.IsNil))

			// Decoders are loaded
			c.Expect(len(pipeConfig.DecoderWrappers), gs.Equals, 2)
			c.Expect(DecodersByEncoding[message.Header_JSON], gs.Equals, "JsonDecoder")
			c.Expect(DecodersByEncoding[message.Header_PROTOCOL_BUFFER], gs.Equals,
				"ProtobufDecoder")
		})
		c.Specify("explodes w/ bad config file", func() {
			err := pipeConfig.LoadFromConfigFile("../testsupport/config_bad_test.toml")
			c.Assume(err, gs.Not(gs.IsNil))
			c.Expect(err.Error(), ts.StringContains, "2 errors loading plugins")
			c.Expect(pipeConfig.logMsgs, gs.ContainsAny, gs.Values("No such plugin: CounterOutput"))
		})

		c.Specify("handles missing config file correctly", func() {
			err := pipeConfig.LoadFromConfigFile("no_such_file.toml")
			c.Assume(err, gs.Not(gs.IsNil))
			c.Expect(err.Error(), ts.StringContains, "open no_such_file.toml: no such file or directory")
		})

		c.Specify("errors correctly w/ bad outputs config", func() {
			err := pipeConfig.LoadFromConfigFile("../testsupport/config_bad_outputs.toml")
			c.Assume(err, gs.Not(gs.IsNil))
			c.Expect(err.Error(), ts.StringContains, "1 errors loading plugins")
			msg := pipeConfig.logMsgs[0]
			c.Expect(msg, ts.StringContains, "No such plugin")
		})

		c.Specify("captures plugin Init() panics", func() {
			RegisterPlugin("PanicOutput", func() interface{} {
				return new(PanicOutput)
			})
			err := pipeConfig.LoadFromConfigFile("../testsupport/config_panic.toml")
			c.Expect(err, gs.Not(gs.IsNil))
		})
	})
}
