// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package driver

import "context"

// Config represents all the config sections.
type Config map[string]ConfigSection

// ConfigSection represents all key/value pairs for a section of configuration.
type ConfigSection map[string]string

// Configer is an optional interface that may be implemented by a [Client] to
// allow access to reading and setting server configuration.
type Configer interface {
	Config(ctx context.Context, node string) (Config, error)
	ConfigSection(ctx context.Context, node, section string) (ConfigSection, error)
	ConfigValue(ctx context.Context, node, section, key string) (string, error)
	SetConfigValue(ctx context.Context, node, section, key, value string) (string, error)
	DeleteConfigKey(ctx context.Context, node, section, key string) (string, error)
}
