extends Camera

var look_active = false
var view_sensitivity = 0.25
var yaw = 0
var pitch = 0

const CAMERA_SPEED = 1.0

func _ready():
	# Called every time the node is added to the scene.
	# Initialization here
	set_process_input(true)


func _input(event):
	if event.type == InputEvent.MOUSE_BUTTON and event.button_index == 1 and event.pressed:
		look_active = true
		
	if event.type == InputEvent.MOUSE_BUTTON and event.button_index == 1 and not event.pressed:
		look_active = false
		
	if look_active and event.type == InputEvent.MOUSE_MOTION:
			yaw = fmod(yaw - event.relative_x * view_sensitivity, 360)
			pitch = max(min(pitch - event.relative_y * view_sensitivity, 85), -85)
			set_rotation_deg(Vector3(pitch, yaw, 0))
	if Input.is_action_pressed("camera_forward"):
		var t = get_transform()
		var r = get_rotation()
		t.origin += r*CAMERA_SPEED
		set_transform(t)
	if Input.is_action_pressed("camera_backward"):
		var t = get_transform()
		var r = get_rotation()
		t.origin -= r*CAMERA_SPEED
		set_transform(t)
