package pubsub

import (
	// "bytes"
	"log"
	"sync"

	"github.com/puzpuzpuz/xsync/v4"
	// "log"
	// "math"
	// "time"
)

type Room struct {
	lastRoomInfo []byte
	Subscribers  []string // keys to Users xsync.Map
}

type User struct {
	UserId string
	Chs    []chan []byte
}

type xsyncCompute[T any] = func(oldValue *T, loaded bool) (newValue *T, op xsync.ComputeOp)

func (p *PubSub) InitPS(x int) {
	fconf := xsync.WithPresize(x)
	p.Rooms = xsync.NewMap[string, *Room](fconf)
	p.Users = xsync.NewMap[string, *User](fconf)

	p.RoomsPool = sync.Pool{New: func() any {
		r := Room{
			lastRoomInfo: make([]byte, 0),
			Subscribers:  make([]string, 0),
		}
		return &r
	}}

	p.ChanPool = sync.Pool{New: func() any {
		ch := make(chan []byte, 250) /* each connection can have up to 250 rooms
		and up to 10 connections for one user */
		return ch
	}}
}

// subscribe/unsubscirbe users to/from rooms
func (p *PubSub) SubscribeUsers(users []string, room string, subscribe bool, skipIfLoaded bool) {
	cop := p.RoomSubscribeCompute(users, subscribe, skipIfLoaded)
	p.Rooms.Compute(room, cop)
}

// to publish a message in the room
func (p *PubSub) PublishWithMessage(PublishCh string, message []byte) {
	c := p.RoomDispatchMsgCompute(PublishCh, message)
	p.Rooms.Compute(PublishCh, c)
}

func (p *PubSub) GetUser(userid string) chan []byte {
	ch := p.GetChanPool()
	if ch == nil {
		return nil
	}
	notAllowed := false
	c := p.GetUserCompute(userid, ch, &notAllowed)
	p.Users.Compute(userid, c)
	if notAllowed {
		return nil
	}
	return ch
}

func (p *PubSub) ReleaseUser(userid string, ch chan []byte) {
	c := p.ReleaseUserCompute(userid, ch)
	p.Users.Compute(userid, c)
	p.ChanPool.Put(ch)
}

func (p *PubSub) RoomDispatchMsgCompute(PublishCh string, message []byte) xsyncCompute[Room] {
	return func(oldValue *Room, loaded bool) (newValue *Room, op xsync.ComputeOp) {
		cop := xsync.CancelOp
		var rm *Room
		if loaded {
			rm = oldValue
		} else {
			r := p.GetRoomPool()
			if r == nil {
				return nil, cop
			}
			rm = r
			cop = xsync.UpdateOp
		}

		for _, v := range rm.Subscribers {
			c := p.UserDispatchMsgCompute(v, message)
			p.Users.Compute(v, c)
		}
		return rm, cop
	}
}
func (p *PubSub) RoomSubscribeCompute(users []string, subscribe bool, skipIfLoaded bool) xsyncCompute[Room] {
	return func(oldValue *Room, loaded bool) (newValue *Room, op xsync.ComputeOp) {
		cop := xsync.CancelOp

		var rm *Room
		if loaded {
			if skipIfLoaded {
				return nil, cop
			}
			rm = oldValue
		} else {
			r := p.GetRoomPool()
			if r == nil {
				return nil, cop
			}
			rm = r
			cop = xsync.UpdateOp
		}
		if subscribe {
			for i := range users {
			loop:
				for _, v := range rm.Subscribers {
					if v == users[i] {
						break loop
					}
				}
				log.Println("added in subs", users[i])
				rm.Subscribers = append(rm.Subscribers, users[i])
			}
		} else {
			for _, usr := range users {
				for i, sub := range rm.Subscribers {
					if usr == sub {
						lenr := len(rm.Subscribers) - 1
						lastsub := rm.Subscribers[lenr]
						rm.Subscribers[i] = lastsub
						rm.Subscribers = rm.Subscribers[:lenr]
					}
				}
			}
			if len(rm.Subscribers) == 0 {
				p.RoomsPool.Put(rm)
				return nil, xsync.DeleteOp
			}
		}
		return rm, cop
	}
}

func (p *PubSub) UserDispatchMsgCompute(userid string, message []byte) xsyncCompute[User] {
	return func(oldValue *User, loaded bool) (newValue *User, op xsync.ComputeOp) {
		cop := xsync.CancelOp
		var usr *User
		if loaded {
			usr = oldValue
		} else {
			usr = &User{UserId: userid, Chs: make([]chan []byte, 0)}
			cop = xsync.UpdateOp
		}
		for i := range usr.Chs {
			log.Println(i, "chan", usr.UserId)

			usr.Chs[i] <- message
		}
		return usr, cop
	}
}

func (p *PubSub) GetRoomPool() *Room {
	if r, ok := p.RoomsPool.Get().(*Room); ok {
		r.lastRoomInfo = r.lastRoomInfo[:0]
		r.Subscribers = r.Subscribers[:0]
		return r
	}
	return nil
}
func (p *PubSub) GetChanPool() chan []byte {
	if ch, ok := p.ChanPool.Get().(chan []byte); ok {
		for range len(ch) {
			<-ch
		}
		return ch
	}
	return nil
}
func (p *PubSub) GetUserCompute(userid string, ch chan []byte, notAllowed *bool) xsyncCompute[User] {
	return func(oldValue *User, loaded bool) (newValue *User, op xsync.ComputeOp) {

		var usr *User
		cop := xsync.CancelOp
		if loaded {
			usr = oldValue
		} else {
			usr = &User{UserId: userid, Chs: make([]chan []byte, 0)}
			cop = xsync.UpdateOp
		}
		if len(usr.Chs) > 10 {
			*notAllowed = true
			return nil, xsync.CancelOp
		}
		usr.Chs = append(usr.Chs, ch)
		return usr, cop
	}
}
func (p *PubSub) ReleaseUserCompute(userid string, ch chan []byte) xsyncCompute[User] {
	return func(oldValue *User, loaded bool) (newValue *User, op xsync.ComputeOp) {

		var usr *User
		cop := xsync.CancelOp
		if loaded {
			usr = oldValue
		} else {
			usr = &User{UserId: userid, Chs: make([]chan []byte, 0)}
			cop = xsync.UpdateOp
		}
		for i, chu := range usr.Chs {
			if chu == ch {
				lench := len(usr.Chs) - 1
				lastch := usr.Chs[lench]
				usr.Chs[i] = lastch
				usr.Chs = usr.Chs[:lench]
			}
		}
		return usr, cop
	}
}
