extends Control

onready var _error_msg = get_node("error_msg")

func _ready():
	# Called every time the node is added to the scene.
	# Initialization here
	set_process(true)
	
func _process(delta):
	_error_msg.set_text(global.error_msg)
