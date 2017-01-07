extends Camera

const CAMERA_SPEED = 2.0

func _ready():
	# Called every time the node is added to the scene.
	# Initialization here
	set_process_input(true)
	set_process(true)

var value = 0
var srcTrans = null
var dstTrans = null
var looking = false

func _process(delta):
	if looking:
		value += delta * CAMERA_SPEED
		if value > 1:
			value = 1 
			looking = false
		var smooth = value*value*(3-2*value)
		var thisRotation = Quat(srcTrans.basis).slerpni(dstTrans.basis,smooth)
		set_transform(Transform(thisRotation, get_camera_transform().origin))
		if !looking:
			value = 0
	

#Input handler, listen for ESC to exit app
func _input(event):
	if event.type == InputEvent.MOUSE_BUTTON and event.button_index == 1 and event.pressed:
		var mousePos = get_viewport().get_mouse_pos()
		var lookDir = project_position(mousePos)
		#look_at(lookDir, Vector3(0,1,0))
		
		var curTrans = get_camera_transform()
		var rotTrans = curTrans.looking_at(lookDir, Vector3(0,1,0))
		srcTrans = curTrans
		dstTrans = rotTrans
		looking = true
	if event.type == InputEvent.MOUSE_BUTTON and event.is_pressed() and not event.is_echo() and (event.button_index == BUTTON_WHEEL_DOWN or event.button_index == BUTTON_WHEEL_UP):
		var t = get_camera_transform()
		var rm = t.basis
		var diff = rm.xform(Vector3(0,0,1))
		if event.button_index == BUTTON_WHEEL_DOWN:
			set_translation(get_translation() + diff)
		else:
			set_translation(get_translation() - diff)