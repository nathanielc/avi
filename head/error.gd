extends Reference

var _message = ""
var _value = null
var _cause = null

func _ready():
	pass

func _init(m, v, c):
	_message = m
	_value = v
	_cause = c

func is_ok():
	return _message == ""

func message():
	var msg = _message
	if _cause != null:
		msg += ": " + _cause.message()
	return msg

func value():
	return _value

func get_cause():
	if _cause != null:
		return _cause.get_cause()
	return self