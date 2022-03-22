var Mouse = Java.type('org.lwjgl.input.Mouse')
var TexteriaGui = Java.type('net.xtrafrancyz.mods.texteria.gui.TexteriaGui')

var dragging = false

function _click(x, y, button) {
    dragging = button === 0 && self.hover
}

function _beforeRender() {
    if (dragging && !self.hover) dragging = false
    if (!Mouse.isButtonDown(0) || !dragging) return
    var scale = 1 / TexteriaGui.scaledResolution.e()
    self.x.set(self.x.render + Mouse.getDX() * scale)
    self.y.set(self.y.render - Mouse.getDY() * scale)
}
