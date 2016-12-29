extends Control

func _ready():
	get_node("watch").connect("pressed", self, "_on_watch_pressed")
	var r = global.client.get_games()
	if r.has('games') and r['games'] != null:
		var games = r['games']
		for gid in games.keys():
			var g = games[gid]
			var live = 'live'
			if not g['active']:
				live = 'saved'
			get_node("games").add_item(g['date'].substr(0,16) + " " + g['id'] + " " + live)
		get_node("games").select(0)
		get_node("games").sort_items_by_text()

func _on_watch_pressed():
	var game_id = get_node("games").get_item_text(get_node("games").get_selected_items()[0]).split(" ")[1]
	
	global.game_id = game_id
	global.goto_scene("res://ui/views/game.tscn")