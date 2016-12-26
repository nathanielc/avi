extends Control

# class member variables go here, for example:
# var a = 2
# var b = "textvar"

func _ready():
	# Called every time the node is added to the scene.
	# Initialization here
	get_node("start").connect("pressed", self, "_on_start_pressed")
	var maps = global.client.get_maps()
	for name in maps.keys():
		get_node("maps").add_item(name)
	get_node("maps").select(0)
	get_node("maps").sort_items_by_text()
		
	var part_sets = global.client.get_part_sets()
	for name in part_sets.keys():
		get_node("part_sets").add_item(name)
	get_node("part_sets").select(0)
	get_node("part_sets").sort_items_by_text()
		
	var fleets = global.client.get_fleets()
	for name in fleets.keys():
		get_node("fleets").add_item(name)
	get_node("fleets").set_select_mode(ItemList.SELECT_MULTI)
	get_node("fleets").sort_items_by_text()
	get_node("fleets").select(0,false)
	get_node("fleets").select(1,false)


func _on_start_pressed():
	var map = get_node("maps").get_item_text(get_node("maps").get_selected_items()[0])
	var part_set = get_node("part_sets").get_item_text(get_node("part_sets").get_selected_items()[0])
	
	var fleets = []
	for i in get_node("fleets").get_selected_items():
		fleets.append(get_node("fleets").get_item_text(i))
	
	global.game_id = global.client.start_game(map, part_set, fleets)
	global.goto_scene("res://ui/views/game.tscn")
