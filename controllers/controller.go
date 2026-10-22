package controllers

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type pagingResult struct {
	Page      int `json:"page"`
	PrevPage  int `json:"prevPage"`
	NextPage  int `json:"nextPage"`
	Limit     int `json:"limit"`
	Count     int `json:"count"`
	TotalPage int `json:"totalPage"`
}

type pagination struct {
	ctx     *gin.Context
	query   *gorm.DB
	records interface{}
}

func (p *pagination) paginate() *pagingResult {
	page, _ := strconv.Atoi(p.ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(p.ctx.DefaultQuery("limit", "12"))

	/*
		count records
		 ต้องการให้ query.Model กับ query.Limit เข้าถึงฐานข้อมูลพร้อมกันได้
		 จึงต้องแปลงการทำงานของการนับ count เป็น go routine
	*/

	// 2. Count records
	ch := make(chan int)
	go p.countRecords(ch)

	count := <-ch
	// 3.Find Record (Limit, Offset)
	offset := (page - 1) * limit
	p.query.Limit(limit).Offset(offset).Find(p.records)

	// 4. Total Page
	totalPage := int(math.Ceil(float64(count) / float64(limit))) // จำนวนทั้งหมด หาร limit แล้วปัดขึ้น

	var nextPage int
	if page == totalPage {
		nextPage = totalPage
	} else {
		nextPage = page + 1
	}

	return &pagingResult{
		Page:      page,
		Limit:     limit,
		Count:     int(count),
		PrevPage:  page - 1,
		NextPage:  nextPage,
		TotalPage: totalPage,
	}

}

func (p *pagination) countRecords(ch chan int) {
	var count int64
	p.query.Model(p.records).Count(&count)

	ch <- int(count)
}
