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

function _interopDefault (ex) { return (ex && (typeof ex === 'object') && 'default' in ex) ? ex['default'] : ex; }
var inherits = _interopDefault(require('inherits'));
inherits(PouchError, Error);

function PouchError(status, error, reason) {
    Error.call(this, reason);
    this.status = status;
    this.name = error;
    this.message = reason;
    this.error = true;
}

PouchError.prototype.toString = function () {
    return JSON.stringify({
        status: this.status,
        name: this.name,
        message: this.message,
        reason: this.reason
    });
};

$global.ReconstitutePouchError = function(str) {
    const o = JSON.parse(str);
    Object.setPrototypeOf(o, PouchError.prototype);
    return o;
};
