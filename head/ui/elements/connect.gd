extends Control

const retry_interval = 1

var client = null
var time = retry_interval

var _connected_tex = load("res://ui/images/gear_connected.png")
var _disconnected_tex = load("res://ui/images/gear_disconnected.png")

onready var _configure = get_node("configure")
onready var _status = get_node("status")

func _ready():
	_configure.connect("pressed", self, "_on_congfigure_pressed")
	set_process(true)

func _connect():
	client = preload("res://client.gd").new()
	var err = client.connect("127.0.0.1", 4242)
	if !err.is_ok():
		_status.set_text("Disconnected")
		_configure.set_tooltip(err.message())
		_configure.set_normal_texture(_disconnected_tex)
		return
	_status.set_text("Connected")
	_configure.set_normal_texture(_connected_tex)
	_configure.set_tooltip("")
	global.set_client(client)

func _process(delta):
	if client == null or !client.is_connected():
		time += delta
		if time > retry_interval:
			time = 0
			_connect()

func _on_congfigure_pressed():
	pass
