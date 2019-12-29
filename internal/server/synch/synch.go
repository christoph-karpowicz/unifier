package synch

import (
	"log"

	"github.com/christoph-karpowicz/unifier/internal/server/db"
)

type synch struct {
	synch   *synchData
	dbs     map[string]*db.Database
	tables  map[string]*table
	running bool
	initial bool
}

func (s *synch) GetData() *synchData {
	return s.synch
}

func (s *synch) Init(DBMap map[string]*db.Database) {
	s.tables = make(map[string]*table)
	s.setDatabases(DBMap)
	s.copyTables()
	s.assignTablePointers()
	s.selectData()
	s.pairData()
	s.setParentPointers()
}

func (s *synch) assignTablePointers() {
	for i := range s.synch.Vectors {
		var vctr *vector = &s.synch.Vectors[i]

		vctr.sourceTable = s.tables[vctr.Source.Table]
		vctr.targetTable = s.tables[vctr.Target.Table]
	}
}

func (s *synch) copyTable(endpoint vectorEndpoint) {
	_, tableCopied := s.tables[endpoint.Database+"."+endpoint.Table]
	if !tableCopied {
		tbl := &table{
			id:     endpoint.Database + "." + endpoint.Table,
			dbName: endpoint.Database,
			name:   endpoint.Table,
		}
		rawRecords := (*s.dbs[endpoint.Database]).Select(tbl.name, "-")

		if !s.initial {
			tbl.oldRecords = tbl.records
		}

		tbl.records = &tableRecords{records: mapToRecords(rawRecords, endpoint.Key)}
		s.tables[tbl.id] = tbl
	}
}

func (s *synch) copyTables() {
	for i := range s.synch.Vectors {
		var vctr *vector = &s.synch.Vectors[i]

		s.copyTable(vctr.Source)
		s.copyTable(vctr.Target)
	}
}

// pairData pairs together records that are going to be synchronized.
func (s *synch) pairData() {
	for i := range s.synch.Vectors {
		var vector *vector = &s.synch.Vectors[i]
		vector.createPairs()
	}
}

// Selects all records from all tables and filters them to get the relevant records.
func (s *synch) selectData() {
	for i := range s.synch.Vectors {
		var vctr *vector = &s.synch.Vectors[i]
		sourceRawActiveRecords := (*s.dbs[vctr.Source.Database]).Select(vctr.Source.Table, vctr.Source.Condition)
		targetRawActiveRecords := (*s.dbs[vctr.Target.Database]).Select(vctr.Target.Table, vctr.Target.Condition)

		if !s.initial {
			vctr.sourceOldActiveRecords = vctr.sourceActiveRecords
			vctr.targetOldActiveRecords = vctr.targetActiveRecords
		}

		for _, sourceRecord := range sourceRawActiveRecords {
			sourceRecordPointer := vctr.sourceTable.records.FindRecordPointer(sourceRecord)
			vctr.sourceActiveRecords = append(vctr.sourceActiveRecords, sourceRecordPointer)
			sourceRecordPointer.ActiveIn = append(sourceRecordPointer.ActiveIn, vctr)
		}
		for _, targetRecord := range targetRawActiveRecords {
			targetRecordPointer := vctr.targetTable.records.FindRecordPointer(targetRecord)
			vctr.targetActiveRecords = append(vctr.targetActiveRecords, targetRecordPointer)
			targetRecordPointer.ActiveIn = append(targetRecordPointer.ActiveIn, vctr)
		}
		log.Println(vctr.sourceActiveRecords)
		log.Println(vctr.targetActiveRecords)
	}
}

func (s *synch) setDatabase(DBMap map[string]*db.Database, endpoint vectorEndpoint) {
	_, dbExists := DBMap[endpoint.Database]
	if dbExists {
		s.dbs[endpoint.Database] = DBMap[endpoint.Database]
		(*s.dbs[endpoint.Database]).Init()
	} else {
		panic("Database " + endpoint.Database + " hasn't been configured.")
	}
}

// Open chosen database connections.
func (s *synch) setDatabases(DBMap map[string]*db.Database) {
	for j := range s.synch.Vectors {
		var vctr *vector = &s.synch.Vectors[j]

		s.setDatabase(DBMap, vctr.Source)
		s.setDatabase(DBMap, vctr.Target)
	}
}

func (s *synch) setParentPointers() {
	for j := range s.synch.Vectors {
		var vector *vector = &s.synch.Vectors[j]
		// vector.table = table

		for k := range vector.pairs {
			var pair *pair = &s.synch.Vectors[j].pairs[k]
			pair.vector = vector
		}
	}
}

func (s *synch) SynchPairs() {
	for j := range s.synch.Vectors {
		var vctr *vector = &s.synch.Vectors[j]

		for k := range vctr.pairs {
			var pair *pair = &vctr.pairs[k]
			_, err := pair.synchronize(s.dbs[vctr.Source.Database], s.dbs[vctr.Target.Database])
			if err != nil {
				log.Println(err)
			}
		}
	}
}
