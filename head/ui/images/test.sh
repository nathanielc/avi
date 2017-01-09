#!/bin/bash



game_id=$(curl -X POST  http://localhost:4242/avi/games --data-binary '{"map":"arden","part_set":"arden","fleets":["nathanielc","DubberHeads"]}' | jq .id -r )

sleep 0.1

curl -i "http://localhost:4242/avi/games/${game_id}?start=0&stop=100" > test.log

