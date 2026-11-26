package handlers

import (
	"github.com/gin-gonic/gin"
)

func GetProfile(c *gin.Context) {
	getDefaultUserHandler().GetProfile(c)
}

func UpdateProfile(c *gin.Context) {
	getDefaultUserHandler().UpdateProfile(c)
}

func GetDirectory(c *gin.Context) {
	getDefaultUserHandler().GetDirectory(c)
}

func TeacherUpdateUserRole(c *gin.Context) {
	getDefaultUserHandler().TeacherUpdateUserRole(c)
}

func TeacherCreateClass(c *gin.Context) {
	getDefaultClassroomHandler().TeacherCreateClass(c)
}

func TeacherUpdateClass(c *gin.Context) {
	getDefaultClassroomHandler().TeacherUpdateClass(c)
}

func TeacherDeleteClass(c *gin.Context) {
	getDefaultClassroomHandler().TeacherDeleteClass(c)
}

func TeacherAddClassMember(c *gin.Context) {
	getDefaultClassroomHandler().TeacherAddClassMember(c)
}

func TeacherRemoveClassMember(c *gin.Context) {
	getDefaultClassroomHandler().TeacherRemoveClassMember(c)
}

func StudentJoinClass(c *gin.Context) {
	getDefaultClassroomHandler().StudentJoinClass(c)
}

func ListClasses(c *gin.Context) {
	getDefaultClassroomHandler().ListClasses(c)
}

func GetClassDetail(c *gin.Context) {
	getDefaultClassroomHandler().GetClassDetail(c)
}

func TeacherGetClassGradebook(c *gin.Context) {
	getDefaultClassroomHandler().TeacherGetClassGradebook(c)
}
