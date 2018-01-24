

# Halite2

## Building

go get github.com/fogleman/gg
go get github.com/cevaris/ordered_map

## Acknowledgements

Fohristiwhirl. He wrote chlorine and the replayer, which I relied on heavily for this competition. He also graciously provided me with the starter kit for Go, which I had never written in before. Thank you so much, Fohristiwhirl!

## Background


## Priorities

* Rush
* Desertion
* Combat
* Objectives
* EnsureExpansion


## Rushes
I played with some meta-game things like bait with one and try to snipe with two but these were too driven by enemy prediction. I tried also to snipe with three (check out Fohrisitwhirl's write up if you haven't already - https://github.com/fohristiwhirl/halite2_rush_theory) but I was limited by my navigation at the time. I should've tried again later, but didn't think of it.

## Desertion
I hate this. I hate that this is incentivized. I had to do something for this, but I didn't want to. I decided if I have half of the ships of the largest player (and largest > 40) just get out. I hade two modes of retreat based on what place I was going to be in: turtle all my ships into a corner or circle the map.

## Combat
Open space combat I think I did particularly well in.

Every ship in range of its enemy has the list of combatants (enemy ships within potential damage range) written onto it. I attacked an enemy he had more combatants than all of his combatants (i.e. if I had more ships to attack him than he had to attack any of my ships). If so, I moved the closest combatant in towards him and all the rest of the combtants I move to flanking positions (drawing a right angle from the enemy, through the lead, and to the flanking positions).

Any remaining ships with combatants are those outnumbered. Each of these I retreated to the nearest safe ships. This naturally create clumps out of spread ships that are outnumbered.

BIG problem. Because all combat comes before all objective assignment, ships would often get "stuck" in combat and dragged out of position. Next time I need to integrate combat concerns into my objectives so they can compete rather than just overriding them completely.


## Objective Assignment

Each ship would rank the planets based on distance, size and owner. The ships get assigned to the first in their objectives list. Each planet uses the ships it needs and releases the rest to be assigned to the next objective in their queue. Offensive targets had no limit, defense is trickier. I don't know that I ever came up with a great solution for it, but defensive assignment to a planet was use enough ships for defense first then dock as many as you can.

What I struggled with here was counting the defense. Because I could have ships engaged with enemies near the planets, it wasn't as simple as one per enemy in some range. I defended against enemies five turns out for the surface, but not any further, which really hurt me I think.

Another weakness of this approach is that I would double defend against an enemy that was between two of my planets. In some sense this is good because it kept him from going to either side but on the other hand, it doubled up on defense when I could have positioned on ship better to cover both planets perhaps


## Navigation
I had a more interesting method of navigation, but it wasn't as flexible as I wanted. Eventually I broke down and did what I think was pretty standard. I used the starter bot collision detection, with the tweak for double motion where you just switch into the reference frame of one of the bots. I then generated all nav options up to 90 degrees in each direction and tossed them into a heap with dist to target as the comparison function. I then pulled off the top of the heap until I found one that wasn't blocked by any planets or ships close enough to matter. Simple but effective. I could've made this run much faster, pretty sure, but I was managing memory poorly.

## Harrass
To my knowledge I started this, but someone else would've done it eventually. I got the idea from a probe harrass in SC2. You attack their workers and they either let you or they try to defend. For a long time, people had trouble reacting proportionately to the threat so this would wind up kiting a bunch of ships around while I macro'd up on the other side of the board. Once people caught on to defending minimally, this mattered less.

## EnsureExpansion

This was another witty piece that I was proud of having. The fact that I needed it at all though indicated a flaw in my design. I noted that my bot would get stuck in situations without expanding, usually because it was commiting a lot to defense at a single point. To ensure expansion, I looked at the first objective of each planet that was unmet and would be considered expansion (called a triple in the code). I compare across my ships to see who has the best score for an expansion and remove it from its current objective. This would cause some inconsistent behavior though as new ships spawned, instad of just comparing it to existing ships, I compared it to a dummy ship that could spawn at any of my planets. This way I would wait for a new ship rather than send on cross map.
