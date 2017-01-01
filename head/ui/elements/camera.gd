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

#	var lookDir = get_node(lookTarget).get_transform().origin - t.origin 
#	var rotTransform = t.looking_at(lookDir,Vector3(0,1,0))
#	var thisRotation = Quat(t.basis).slerp(rotTransform.basis,value)
#	value += delta
#	if value>1:
#    	value = 1
#	set_transform(Transform(thisRotation,t.origin))
#	
	
	
	#var mouse_pos = 
	
	
	
#	if event.type == InputEvent.MOUSE_BUTTON and event.button_index == 1 and event.pressed:
#		look_active = true
#		
#	if event.type == InputEvent.MOUSE_BUTTON and event.button_index == 1 and not event.pressed:
#		look_active = false
#		
#	if look_active and event.type == InputEvent.MOUSE_MOTION:
#			yaw = fmod(yaw - event.relative_x * view_sensitivity, 360)
#			pitch = max(min(pitch - event.relative_y * view_sensitivity, 85), -85)
#			set_rotation_deg(Vector3(pitch, yaw, 0))
#	if Input.is_action_pressed("camera_forward"):
#		var t = get_transform()
#		var r = get_rotation()
#		t.origin += r*CAMERA_SPEED
#		set_transform(t)
#	if Input.is_action_pressed("camera_backward"):
#		var t = get_transform()
#		var r = get_rotation()
#		t.origin -= r*CAMERA_SPEED
#		set_transform(t)

	
	
	
	
	
#	if(event.is_pressed()):
#		if(Input.is_key_pressed(KEY_ESCAPE)):
#			get_tree().quit()
#	
#	var zoom = get_node("Camera").get_zoom()
#	
#	if (event.type == InputEvent.MOUSE_MOTION):
#		if(drag == true):
#	
#			var mouse_pos = get_global_mouse_pos()
#	
#			var dist_x = initPosMouse.x - mouse_pos.x
#			var dist_y = initPosMouse.y - mouse_pos.y
#	
#			var nx = initPosNode.x - (0 + dist_x)
#			var ny = initPosNode.y - (0 + dist_y)
#	
#			get_node("hud").set_pos(Vector2(nx,ny))
#	
#		elif(drag == false):
#			# print("undrag")
#			pass
#	
#	if (event.type == InputEvent.MOUSE_BUTTON):
#		if (event.button_index == BUTTON_WHEEL_UP):
#			# print("wheel up (event)")
#			zoom[0] = zoom[0] + 0.25
#			zoom[1] = zoom[1] + 0.25
#		if (event.button_index == BUTTON_WHEEL_DOWN):
#			# print("wheel down (event)")
#			if(zoom[0] - 0.25 > 0 && zoom[1] - 0.25 > 0):
#				zoom[0] = zoom[0] - 0.25
#				zoom[1] = zoom[1] - 0.25
#		if (event.button_index == BUTTON_MIDDLE):
#			if(Input.is_mouse_button_pressed(3)):
#				print("button middle")
#				initPosMouse = get_global_mouse_pos()
#				initPosNode = get_node("hud").get_pos()
#				drag = true
#			else:
#				print("button middle release")
#				drag = false
#		get_node("Camera").set_zoom(zoom)
#


