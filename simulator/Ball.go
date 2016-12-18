package main

import (
	bs "../proto"

	"math/rand"
)

type Ball struct {
	Id     string
	Xpos   int
	Ypos   int
	Xspeed int
	Yspeed int
}

func NewBall(id string) *Ball {
	// just initialize with some random values
	var b Ball
	b.Id = id
	b.Xpos = rand.Intn(BOARD_SIZE - 2*BALL_RADIUS)
	b.Ypos = rand.Intn(BOARD_SIZE - 2*BALL_RADIUS)
	b.Xspeed = rand.Intn(2*BALL_RADIUS) - BALL_RADIUS
	b.Yspeed = rand.Intn(2*BALL_RADIUS) - BALL_RADIUS
	return &b
}

func (b *Ball) UpdatePosition() {
	// only worry about collisions with the walls, not other balls
	b.Xpos += b.Xspeed
	if b.Xpos+BALL_RADIUS > BOARD_SIZE {
		b.Xpos = BOARD_SIZE - BALL_RADIUS
		b.Xspeed *= -1
	} else if b.Xpos < BALL_RADIUS {
		b.Xpos = BALL_RADIUS
		b.Xspeed *= -1
	}

	b.Ypos += b.Yspeed
	if b.Ypos+BALL_RADIUS > BOARD_SIZE {
		b.Ypos = BOARD_SIZE - BALL_RADIUS
		b.Yspeed *= -1
	} else if b.Ypos < BALL_RADIUS {
		b.Ypos = BALL_RADIUS
		b.Yspeed *= -1
	}
}

func (b *Ball) ToProto() *bs.BallState {
	return &bs.BallState{Xpos: int32(b.Xpos), Ypos: int32(b.Ypos), Id: b.Id}
}
