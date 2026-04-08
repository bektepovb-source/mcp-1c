package dump

// MetadataType describes a single 1C metadata kind with its English singular
// name (used in tool input), English plural (the dump directory name), and
// the Russian singular display name.
type MetadataType struct {
	SingularEng string // e.g. "Catalog"
	PluralEng   string // e.g. "Catalogs"   (dump directory)
	RussianName string // e.g. "Справочник"  (display prefix)
}

// metadataTypes is the single source of truth for all 1C metadata type
// mappings. Both objectTypeToDumpDir and dumpDirNames are derived from it.
var metadataTypes = []MetadataType{
	{"Catalog", "Catalogs", "Справочник"},
	{"Document", "Documents", "Документ"},
	{"DataProcessor", "DataProcessors", "Обработка"},
	{"Report", "Reports", "Отчет"},
	{"InformationRegister", "InformationRegisters", "РегистрСведений"},
	{"AccumulationRegister", "AccumulationRegisters", "РегистрНакопления"},
	{"AccountingRegister", "AccountingRegisters", "РегистрБухгалтерии"},
	{"CalculationRegister", "CalculationRegisters", "РегистрРасчета"},
	{"ChartOfAccounts", "ChartsOfAccounts", "ПланСчетов"},
	{"ChartOfCharacteristicTypes", "ChartsOfCharacteristicTypes", "ПланВидовХарактеристик"},
	{"ChartOfCalculationTypes", "ChartsOfCalculationTypes", "ПланВидовРасчета"},
	{"ExchangePlan", "ExchangePlans", "ПланОбмена"},
	{"BusinessProcess", "BusinessProcesses", "БизнесПроцесс"},
	{"Task", "Tasks", "Задача"},
	{"Enum", "Enums", "Перечисление"},
	{"Constant", "Constants", "Константа"},
}

// objectTypeToDumpDir maps singular English type name to plural English dump
// directory name (e.g. "Catalog" -> "Catalogs"). Derived from metadataTypes.
var objectTypeToDumpDir map[string]string

// dumpDirNames maps plural English dump directory name to Russian display
// name (e.g. "Catalogs" -> "Справочник"). Derived from metadataTypes.
//
// CommonModules is added separately because it has no singular form used in
// tool input (there is no "CommonModule" object type).
var dumpDirNames map[string]string

func init() {
	objectTypeToDumpDir = make(map[string]string, len(metadataTypes))
	dumpDirNames = make(map[string]string, len(metadataTypes)+1)

	for _, mt := range metadataTypes {
		objectTypeToDumpDir[mt.SingularEng] = mt.PluralEng
		dumpDirNames[mt.PluralEng] = mt.RussianName
	}

	// CommonModules has no corresponding singular object type. It only
	// appears as a dump directory, so we add it to dumpDirNames directly.
	dumpDirNames["CommonModules"] = "ОбщийМодуль"
}
