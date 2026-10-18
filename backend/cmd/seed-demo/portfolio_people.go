package main

import "fmt"

var workerNames = []string{
	"Asep Sunandar", "Ujang Hermawan", "Dedi Kurnia", "Cecep Supriatna", "Wawan Setiadi",
	"Sugeng Riyanto", "Paijo Wiyono", "Tarno Saputro", "Sutrisno Adi", "Margono Hadi",
	"Andi Baso", "Syamsul Rijal", "Haeruddin Lallo", "Rahman Daeng", "Ilham Syarif",
	"Komarudin Saleh", "Endang Suryana", "Iwan Kurniadi", "Nana Sumarna", "Ade Rohendi",
	"Budiyono Slamet", "Gatot Raharjo", "Hartono Wijoyo", "Mulyadi Prasetyo", "Joko Purnomo",
	"Herman Lubis", "Rudi Pangaribuan", "Saiful Anwar", "Yanto Kusnadi", "Agus Salim",
}

var staffNames = []string{
	"Rangga Pradipta", "Sinta Maharani", "Galih Saputra", "Anindya Putri", "Bagas Wicaksono",
	"Cindy Halim", "Dimas Prasetya", "Elsa Natalia", "Farhan Aziz", "Gabriela Sitorus",
	"Hafiz Ramadhan", "Inggrid Tanoto", "Jefri Manurung", "Kirana Larasati", "Leo Hartanto",
	"Mega Kusumawati", "Naufal Hakim", "Olivia Gunawan", "Pandu Wirawan", "Raisa Anjani",
	"Satria Nugroho", "Tiara Ananda", "Umar Faruq", "Vania Salim", "Wahyu Hidayat",
	"Yohana Simatupang", "Zaki Mubarak", "Ayu Wandira", "Bima Arya", "Clara Wibowo",
	"Dewa Putra", "Erlangga Yudha", "Fitri Handayani", "Gilang Ramadhan", "Hesti Purwaningsih",
	"Irma Suryani",
}

type roleSeed struct {
	position   string
	employment string
	salary     int64
	allowance  int64
}

var siteTeam = []roleSeed{
	{"Manajer Proyek", employmentPermanent, 19_500_000, 3_000_000},
	{"Site Engineer", employmentPermanent, 12_000_000, 1_500_000},
	{"Pengawas K3", employmentContract, 9_000_000, 750_000},
	{"Quantity Surveyor", employmentContract, 9_800_000, 750_000},
	{"Admin Proyek", employmentContract, 6_500_000, 500_000},
}

var siteCrew = []roleSeed{
	{"Mandor", employmentDaily, 285_000, 0},
	{"Tukang Besi", employmentDaily, 245_000, 0},
	{"Tukang Kayu", employmentDaily, 240_000, 0},
	{"Tukang Batu", employmentDaily, 235_000, 0},
	{"Pekerja", employmentDaily, 225_000, 0},
}

var officeTeam = []roleSeed{
	{"Manajer SDM", employmentPermanent, 15_000_000, 2_000_000},
	{"Staf Legal & Perizinan", employmentPermanent, 10_500_000, 1_000_000},
	{"Akuntan", employmentPermanent, 11_000_000, 1_000_000},
	{"Staf Pajak", employmentPermanent, 9_000_000, 750_000},
	{"Manajer Pemasaran", employmentPermanent, 16_000_000, 2_500_000},
	{"Staf IT", employmentContract, 9_500_000, 500_000},
}

type staffedProject struct {
	code   string
	joined int
}

func staffedProjects() []staffedProject {
	return []staffedProject{
		{projectSkyGarden, -14}, {projectTembalang, -7}, {projectSchool, -5},
		{projectWarehouse, -9}, {projectOffice, -3}, {projectBuahBatu, -6},
	}
}

func portfolioEmployees() []employeeSeed {
	employees := []employeeSeed{}
	staff, crew := 0, 0
	for _, project := range staffedProjects() {
		for _, role := range siteTeam {
			employees = append(employees, employeeSeed{
				fmt.Sprintf("DEMO-NIK-%03d", 101+len(employees)), staffNames[staff%len(staffNames)], role.position,
				project.code, role.employment, project.joined, role.salary, role.allowance,
			})
			staff++
		}
		for _, role := range siteCrew {
			employees = append(employees, employeeSeed{
				fmt.Sprintf("DEMO-NIK-%03d", 101+len(employees)), workerNames[crew%len(workerNames)], role.position,
				project.code, role.employment, project.joined, role.salary, role.allowance,
			})
			crew++
		}
	}
	for index, role := range officeTeam {
		employees = append(employees, employeeSeed{
			fmt.Sprintf("DEMO-NIK-%03d", 101+len(employees)), staffNames[(staff+index)%len(staffNames)], role.position,
			"", role.employment, -30 + index*3, role.salary, role.allowance,
		})
	}
	employees = append(employees,
		employeeSeed{fmt.Sprintf("DEMO-NIK-%03d", 101+len(employees)), "Kevin Santoso", "Arsitek Perencana", projectMall, employmentPermanent, -5, 14_000_000, 1_500_000},
		employeeSeed{fmt.Sprintf("DEMO-NIK-%03d", 102+len(employees)), "Ratih Puspitasari", "Analis Kelayakan", projectMall, employmentContract, -4, 11_500_000, 750_000},
	)
	return employees
}
