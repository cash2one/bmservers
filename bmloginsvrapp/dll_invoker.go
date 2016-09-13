package main

import (
	"syscall"

	"github.com/cihub/seelog"
)

var (
	g_dllHumSave syscall.DLL
	g_procMap    = make(map[string]*syscall.Proc)

	g_procDllName = []string{
		"CreateHumSave",
		"OpenHumSave",
		"CloseHumSave",
		"AddGameRole",
		"DelGameRole",
		"GetGameRoleInfo_Value",
		"GetGameRoleInfo_Name",
		"GetGameRoleIndex",
		"ReadGameRoleData",
		"WriteGameRoleData",
		"WriteGameRoleInfo",
		"ReadGameRoleSize",
		"UpdateGameRoleInfo",
		"ReadGameRoleHeadInfo",
		"ReadExtendDataSize",
		"ReadExtendData",
		"WriteExtendData",
	}
)

const (
	FUNC_CreateHumSave = iota
	FUNC_OpenHumSave
	FUNC_CloseHumSave
	FUNC_AddGameRole
	FUNC_DelGameRole
	FUNC_GetGameRoleInfo_Value
	FUNC_GetGameRoleInfo_Name
	FUNC_GetGameRoleIndex
	FUNC_ReadGameRoleData
	FUNC_WriteGameRoleData
	FUNC_WriteGameROleInfo
	FUNC_ReadGameRoleSize
	FUNC_UpdateGameRoleInfo
	FUNC_ReadGameRoleHeadInfo
	FUNC_ReadExtendDataSize
	FUNC_ReadExtendData
	FUNC_WriteExtendData
)

func initDllModule(name string) bool {
	allLoaded := true
	g_dllHumSave, err := syscall.LoadDLL(name)
	if err != nil {
		seelog.Error("Can't load [", name, "]")
		return false
	}
	//	Get all module
	for _, str := range g_procDllName {
		proc, err := g_dllHumSave.FindProc(str)
		if err == nil {
			g_procMap[str] = proc
			seelog.Info("Proccess address[", str, "] loaded...")
			//dbgutil.Display("ProcName", str, "ProcAddr", proc)
		} else {
			seelog.Error("ProcName[", str, "] load failed...", err)
			allLoaded = false
		}
	}

	return allLoaded
}

func releaseDllModule() {
	if len(g_dllHumSave.Name) != 0 {
		for _, str := range g_procDllName {
			delete(g_procMap, str)
		}
		seelog.Info(g_dllHumSave.Name, " has been released...")
		g_dllHumSave.Release()
	}
}
