extends Control

func _ready():
	get_node("watch").connect("pressed", self, "_on_watch_pressed")
	var replays = global.client.get_replays()
	print("replays", replays)
	if replays.has('replays') and replays['replays'] != null:
		for r in replays['replays']:
			get_node("replays").add_item(r['date'].substr(0,16) + " " + r['game_id'])
		get_node("replays").select(0)
		get_node("replays").sort_items_by_text()

func _on_watch_pressed():
	var game_id = get_node("replays").get_item_text(get_node("replays").get_selected_items()[0]).split(" ")[1]
	
	global.game_id = game_id
	global.goto_scene("res://ui/views/game.tscn")