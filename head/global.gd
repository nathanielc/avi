extends Node

var client = null
var current_scene = null
var game_id = null
var error_msg = ""
var is_live = false
var conf = ConfigFile.new()
var _conf_path = OS.get_data_dir()+"/avi.ini"

var _error = preload("res://error.gd")

func _ready():
	var root = get_tree().get_root()
	current_scene = root.get_child(root.get_child_count() - 1)
	conf.load(_conf_path)
	
func save():
	conf.save(_conf_path)

func err(msg):
	return _error.new(msg, null, null)
func wrap(err, msg):
	return _error.new(msg, null, err)
func ok(value):
	return _error.new("", value, null)
	
func fail(err):
	set_err(err)
	goto_scene("res://ui/views/main.tscn")
func clear_error():
	global.error_msg = ""
func set_err(err):
	error_msg = "Error message: "+ err.message()

func set_client(c):
	if client != null:
		remove_child(client)
	client = c
	add_child(client)

func goto_scene(path):

	# This function will usually be called from a signal callback,
	# or some other function from the running scene.
	# Deleting the current scene at this point might be
	# a bad idea, because it may be inside of a callback or function of it.
	# The worst case will be a crash or unexpected behavior.

	# The way around this is deferring the load to a later time, when
	# it is ensured that no code from the current scene is running:

	call_deferred("_deferred_goto_scene",path)


func _deferred_goto_scene(path):

	# Immediately free the current scene,
	# there is no risk here.
	current_scene.free()

	# Load new scene
	var s = ResourceLoader.load(path)

	# Instance the new scene
	current_scene = s.instance()

	# Add it to the active scene, as child of root
	get_tree().get_root().add_child(current_scene)

	# optional, to make it compatible with the SceneTree.change_scene() API
	get_tree().set_current_scene(current_scene)
