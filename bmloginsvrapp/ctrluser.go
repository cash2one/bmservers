package main

import (
	"bytes"
	"encoding/binary"

	"github.com/sryanyuan/bmservers/LSControlProto"

	"github.com/axgle/mahonia"
	"github.com/golang/protobuf/proto"
	//	"log"
	"regexp"
	"time"

	"github.com/sryanyuan/bmservers/shareutils"
)

func (this *ServerUser) OnCtrlMsg(msg []byte) {
	buf := bytes.NewBuffer(msg[0:5])
	buf.Next(4)
	var headlen uint8
	binary.Read(buf, binary.LittleEndian, &headlen)
	//	Read head
	head := &LSControlProto.LSCHead{}
	err := proto.Unmarshal(msg[5:5+headlen], head)
	if err != nil {
		shareutils.LogErrorln("Failed to unmarshal proto head")
		return
	}

	//	dispatch
	defer func() {
		except := recover()
		if except != nil {
			shareutils.LogErrorln(except)
		}

		if err != nil {
			shareutils.LogErrorln(err)
		}
	}()

	opcode := LSControlProto.Opcode(head.GetOpcode())
	var oft_body_start int = 5 + int(headlen)
	if opcode != LSControlProto.Opcode_PKG_HeartBeat {
		shareutils.LogDebugln("Ctrl msg[", opcode, "]")
	}

	if !this.ctrlverify {
		//	verify
		if opcode == LSControlProto.Opcode_PKG_CtrlVerifyReq {
			ctrlVerifyReq := &LSControlProto.LSCCtrlVerifyReq{}
			err = proto.Unmarshal(msg[oft_body_start:], ctrlVerifyReq)
			if err != nil {
				shareutils.LogErrorln("proto unmarshal error.", err)
				return
			}

			shareutils.LogInfoln("Verify ctrl terminal[", ctrlVerifyReq.GetVerifycode(), "]")

			ret := &LSControlProto.LSCCtrlVerifyAck{}
			ret.Result = proto.Bool(true)

			if ControlValid(ctrlVerifyReq.GetVerifycode()) {
				this.ctrlverify = true
				shareutils.LogInfoln("pass[", ctrlVerifyReq.GetVerifycode(), "]")
			} else {
				ret.Result = proto.Bool(false)
				shareutils.LogWarnln("invalid terminal[", ctrlVerifyReq.GetVerifycode(), "]")
				//this.conn.GetInternalConn().Close()
			}

			data, _ := proto.Marshal(ret)
			this.SendProtoBuf(uint32(LSControlProto.Opcode_PKG_CtrlVerifyAck), data)
		}
		return
	}

	switch opcode {
	case LSControlProto.Opcode_PKG_RegistAccountReq:
		{
			registAccountReq := &LSControlProto.LSCRegistAccountReq{}
			err = proto.Unmarshal(msg[oft_body_start:], registAccountReq)
			if err != nil {
				return
			}
			this.OnRegistAccountReq(registAccountReq)
		}
	case LSControlProto.Opcode_PKG_RegistAccountWithInfoReq:
		{
			registAccountWithInfoReq := &LSControlProto.RSRegistAccountReq{}
			err = proto.Unmarshal(msg[oft_body_start:], registAccountWithInfoReq)
			if err != nil {
				return
			}
			this.OnRsRegistAccountReq(registAccountWithInfoReq)
		}
	case LSControlProto.Opcode_PKG_ModifyPasswordReq:
		{
			modifyPasswordReq := &LSControlProto.RSModifyPasswordReq{}
			err = proto.Unmarshal(msg[oft_body_start:], modifyPasswordReq)
			if err != nil {
				return
			}
			this.OnRsMofifyPassword(modifyPasswordReq)
		}
	case LSControlProto.Opcode_PKG_InsertDonateRecordReq:
		{
			insertDonateRecordReq := &LSControlProto.RSInsertDonateInfoReq{}
			err = proto.Unmarshal(msg[oft_body_start:], insertDonateRecordReq)
			if err != nil {
				return
			}
			this.OnRsInsertDonateInfoReq(insertDonateRecordReq)
		}
	case LSControlProto.Opcode_PKG_InsertSystemGiftReq:
		{
			insertSystemGiftReq := &LSControlProto.RSInsertSystemGiftReq{}
			err = proto.Unmarshal(msg[oft_body_start:], insertSystemGiftReq)
			if err != nil {
				return
			}
			this.OnRsInsertSystemGiftReq(insertSystemGiftReq)
		}
	}
}

func (this *ServerUser) SendProtoBuf(opcode uint32, msg []byte) bool {
	//	get protobuf head data
	buf := new(bytes.Buffer)
	head := &LSControlProto.LSCHead{}
	head.Opcode = proto.Uint32(opcode)
	headbuf, err := proto.Marshal(head)
	if err != nil {
		shareutils.LogErrorln(err)
		return false
	}

	//	calc size
	var headlength uint8
	headlength = uint8(len(headbuf))

	//	send message
	binary.Write(buf, binary.LittleEndian, &headlength)
	buf.Write(headbuf)
	buf.Write(msg)
	this.conn.Send(buf.Bytes(), 0)

	return true
}

func (this *ServerUser) OnRsRegistAccountReq(req *LSControlProto.RSRegistAccountReq) {
	account := req.GetAccount()
	password := req.GetPassword()

	//	check
	reg, _ := regexp.Compile("^[A-Za-z0-9]+$")
	var ret = false

	if reg.MatchString(account) && reg.MatchString(password) {
		ret = true
	}

	if ret {
		//	regist
		if len(account) > 15 || len(password) > 15 {
			ret = false
		} else {
			users := make([]UserAccountInfo, 1)
			users[0].account = account
			users[0].password = password

			if !dbUserAccountExist(g_DBUser, account) {
				if !dbInsertUserAccountInfo(g_DBUser, users) {
					ret = false
				}
			} else {
				ret = false
			}
		}
	}

	ack := &LSControlProto.RSRegistAccountAck{}
	ack.Result = proto.Bool(ret)
	ack.Account = proto.String(req.GetAccount())
	ack.Mail = proto.String(req.GetMail())
	data, err := proto.Marshal(ack)

	if err != nil {
		shareutils.LogErrorln(err)
		return
	}

	this.SendProtoBuf(uint32(LSControlProto.Opcode_PKG_RegistAccountWithInfoAck),
		data)
}

func (this *ServerUser) OnRegistAccountReq(req *LSControlProto.LSCRegistAccountReq) {
	account := req.GetAccount()
	password := req.GetPassword()

	//	check
	reg, _ := regexp.Compile("^[A-Za-z0-9]+$")
	var ret = false

	if reg.MatchString(account) && reg.MatchString(password) {
		ret = true
	}

	if ret {
		//	regist
		if len(account) > 15 || len(password) > 15 {
			ret = false
		} else {
			users := make([]UserAccountInfo, 1)
			users[0].account = account
			users[0].password = password

			if !dbUserAccountExist(g_DBUser, account) {
				if !dbInsertUserAccountInfo(g_DBUser, users) {
					ret = false
				}
			} else {
				ret = false
			}
		}
	}

	ack := &LSControlProto.LSCRegistAccountAck{}
	ack.Result = proto.Bool(ret)
	data, err := proto.Marshal(ack)

	if err != nil {
		shareutils.LogErrorln(err)
		return
	}

	this.SendProtoBuf(uint32(LSControlProto.Opcode_PKG_RegistAccountAck),
		data)
}

func (this *ServerUser) OnRsMofifyPassword(req *LSControlProto.RSModifyPasswordReq) {
	account := req.GetAccount()
	password := req.GetPassword()

	rsp := &LSControlProto.RSModifyPasswordRsp{}
	rsp.Account = proto.String(account)

	if !dbUserAccountExist(g_DBUser, account) {
		rsp.Result = proto.Bool(false)
		data, err := proto.Marshal(rsp)
		if err != nil {
			shareutils.LogErrorln(err)
			return
		}

		this.SendProtoBuf(uint32(LSControlProto.Opcode_PKG_ModifyPasswordRsp), data)
		return
	}

	if !dbUpdateUserAccountPassword(g_DBUser, account, password) {
		rsp.Result = proto.Bool(false)
		data, err := proto.Marshal(rsp)
		if err != nil {
			shareutils.LogErrorln(err)
			return
		}

		this.SendProtoBuf(uint32(LSControlProto.Opcode_PKG_ModifyPasswordRsp), data)
		return
	}

	rsp.Result = proto.Bool(true)
	data, err := proto.Marshal(rsp)
	if err != nil {
		shareutils.LogErrorln(err)
		return
	}

	this.SendProtoBuf(uint32(LSControlProto.Opcode_PKG_ModifyPasswordRsp), data)
	return
}

func (this *ServerUser) OnRsInsertDonateInfoReq(req *LSControlProto.RSInsertDonateInfoReq) {
	name := req.GetName()
	orderid := req.GetDonateorderid()
	donateMoney := req.GetDonate()

	rsp := &LSControlProto.RSInsertDonateInfoRsp{}
	rsp.Name = proto.String(name)

	//	convert to gbk-name
	enc := mahonia.NewEncoder("gbk")
	playerName := enc.ConvertString(name)

	//	get uid
	uid := dbGetUserUidByName(g_DBUser, playerName)
	if 0 == uid {
		rsp.Result = proto.Int32(-1)
		data, _ := proto.Marshal(rsp)
		this.SendProtoBuf(uint32(LSControlProto.Opcode_PKG_InsertDonateRecordRsp), data)
		return
	}

	//	insert a record
	if !dbIncUserDonateInfo(g_DBUser, uid, int(donateMoney), orderid) {
		rsp.Result = proto.Int32(-2)
		data, _ := proto.Marshal(rsp)
		this.SendProtoBuf(uint32(LSControlProto.Opcode_PKG_InsertDonateRecordRsp), data)
		return
	}

	rsp.Result = proto.Int32(0)
	data, _ := proto.Marshal(rsp)
	this.SendProtoBuf(uint32(LSControlProto.Opcode_PKG_InsertDonateRecordRsp), data)
}

func (this *ServerUser) OnRsInsertSystemGiftReq(req *LSControlProto.RSInsertSystemGiftReq) {
	account := req.GetAccount()
	giftId := req.GetGiftid()
	giftSum := req.GetGiftsum()
	expireTime := req.GetExpiretime()

	rsp := &LSControlProto.RSInsertSystemGiftRsp{}
	rsp.Account = proto.String(account)
	rsp.Result = proto.Int32(0)

	//	get uid
	uid := dbGetUserUidByAccount(g_DBUser, account)
	if 0 == uid {
		rsp.Result = proto.Int32(-1)
		data, _ := proto.Marshal(rsp)
		this.SendProtoBuf(uint32(LSControlProto.Opcode_PKG_InsertSystemGiftRsp), data)
		return
	}

	gift := &SystemGift{}
	gift.uid = uid
	gift.expiretime = int64(expireTime)
	gift.giftid = int(giftId)
	gift.giftsum = int(giftSum)
	gift.givetime = time.Now().Unix()

	// insert a record
	if !dbInsertSystemGift(g_DBUser, gift) {
		rsp.Result = proto.Int32(-2)
		data, _ := proto.Marshal(rsp)
		this.SendProtoBuf(uint32(LSControlProto.Opcode_PKG_InsertSystemGiftRsp), data)
		return
	}

	rsp.Result = proto.Int32(0)
	data, _ := proto.Marshal(rsp)
	this.SendProtoBuf(uint32(LSControlProto.Opcode_PKG_InsertSystemGiftRsp), data)
}
