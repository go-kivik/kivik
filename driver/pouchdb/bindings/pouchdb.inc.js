if ( $global.PouchDB === undefined ) {
    try {
        $global.PouchDB = require('pouchdb');
    } catch(e) {
        throw("kivik: pouchdb bindings: Cannot find global PouchDB object. Did you load the PouchDB library?");
    }
}
try {
    require('pouchdb-all-dbs')($global.PouchDB);
} catch(e) {}
