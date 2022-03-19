var Mouse = Java.type('org.lwjgl.input.Mouse')
var TexteriaGui = Java.type('net.xtrafrancyz.mods.texteria.gui.TexteriaGui')

function _beforeRender() {
    if (!Mouse.isButtonDown(0) || !self.hover) return
    var scale = 1 / TexteriaGui.scaledResolution.e()
    self.x.set(self.x.render + Mouse.getDX() * scale)
    self.y.set(self.y.render - Mouse.getDY() * scale)
}
