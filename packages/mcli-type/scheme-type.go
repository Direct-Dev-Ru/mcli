package mclitype

type StoreType int

const (
	StoreTypeRedis StoreType = iota + 1
	StoreTypePGSql
	StoreTypeFile
)

type PKType int

const (
	PKTypeSequence PKType = iota + 1
	PKTypeFieldValue
	PKTypeGuid
)

type RecordType int

const (
	RecordTypePlain RecordType = iota + 1
	RecordTypeHashTable
	RecordTypeSet
)

type SchemeIndex struct {
	IndexName string
	Fields    []string
	NotUnique bool
}

type Scheme struct {
	StoreType      StoreType
	PKType         PKType
	PKFieldName    string
	Indexes        []SchemeIndex
	RecordType     RecordType
	Prefix         string
	encyptedFields []string
	version        string
}

func NewScheme(storeType StoreType, ver string) *Scheme {
	if storeType == 0 {
		storeType = StoreTypeRedis
	}
	scheme := &Scheme{StoreType: storeType}
	return scheme
}

func (sch *Scheme) SetEncryptedFields(ef []string) {
	sch.encyptedFields = ef
}

func (sch *Scheme) SetSchemeVersion(version string) {
	sch.version = version
}

func (sch *Scheme) GetSchemeVersion() string {
	return sch.version
}
