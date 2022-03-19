function _onStateChanged(pressed) {
    if (!pressed) return
    var ByteMap = Java.type('net.xtrafrancyz.util.ByteMap')
    var Texteria = Java.type('net.xtrafrancyz.mods.texteria.Texteria')

    var map = new ByteMap()
    map.put('%', 'kv:module:toggle')
    map.put('module', '{{ .ModuleIdentifier }}')
    Texteria.sendPacket(map)
}
