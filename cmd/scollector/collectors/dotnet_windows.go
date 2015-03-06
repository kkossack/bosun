package collectors

import (
	"regexp"
	"strings"

	"bosun.org/metadata"
	"bosun.org/opentsdb"
)

var regexesDotNet = []*regexp.Regexp{}

func init() {
	AddProcessDotNetConfig = func(line string) error {
		reg, err := regexp.Compile(line)
		if err != nil {
			return err
		}
		regexesDotNet = append(regexesDotNet, reg)
		return nil
	}
	WatchProcessesDotNet = func() {
		if len(regexesDotNet) == 0 {
			// if no processesDotNet configured in config file, use this set instead.
			regexesDotNet = append(regexesDotNet, regexp.MustCompile("^w3wp"))
		}
		c := &IntervalCollector{
			F: c_dotnet_loading,
		}
		c.init = wmiInit(c, func() interface{} {
			return &[]Win32_PerfRawData_NETFramework_NETCLRLoading{}
		}, "", &dotnetLoadingQuery)
		collectors = append(collectors, c)
	}
}

var (
	dotnetLoadingQuery string
)

func c_dotnet_loading() (opentsdb.MultiDataPoint, error) {
	var dst []Win32_PerfRawData_NETFramework_NETCLRLoading
	err := queryWmi(dotnetLoadingQuery, &dst)
	if err != nil {
		return nil, err
	}
	var md opentsdb.MultiDataPoint
	for _, v := range dst {
		if !nameMatches(v.Name, regexesDotNet) {
			continue
		}
		id := "0"
		raw_name := strings.Split(v.Name, "#")
		name := raw_name[0]
		if len(raw_name) == 2 {
			id = raw_name[1]
		}
		// If you have a hash sign in your process name you don't deserve monitoring ;-)
		if len(raw_name) > 2 {
			continue
		}
		tags := opentsdb.TagSet{"name": name, "id": id}
		Add(&md, "dotnet.current.appdomains", v.Currentappdomains, tags, metadata.Gauge, metadata.Count, descWinDotNetLoadingCurrentappdomains)
		Add(&md, "dotnet.current.assemblies", v.CurrentAssemblies, tags, metadata.Gauge, metadata.Count, descWinDotNetLoadingCurrentAssemblies)
		Add(&md, "dotnet.current.classes", v.CurrentClassesLoaded, tags, metadata.Gauge, metadata.Count, descWinDotNetLoadingCurrentClassesLoaded)
		Add(&md, "dotnet.total.appdomains", v.TotalAppdomains, tags, metadata.Gauge, metadata.Count, descWinDotNetLoadingTotalAppdomains)
		Add(&md, "dotnet.total.appdomains_unloaded", v.Totalappdomainsunloaded, tags, metadata.Gauge, metadata.Count, descWinDotNetLoadingTotalappdomainsunloaded)
		Add(&md, "dotnet.total.assemblies", v.TotalAssemblies, tags, metadata.Gauge, metadata.Count, descWinDotNetLoadingTotalAssemblies)
		Add(&md, "dotnet.total.classes", v.TotalClassesLoaded, tags, metadata.Gauge, metadata.Count, descWinDotNetLoadingTotalClassesLoaded)
		Add(&md, "dotnet.total.load_failures", v.TotalNumberofLoadFailures, tags, metadata.Gauge, metadata.Count, descWinDotNetLoadingTotalNumberofLoadFailures)
	}
	return md, nil
}

const (
	descWinDotNetLoadingCurrentappdomains         = "This counter displays the current number of AppDomains loaded in this application. AppDomains (application domains) provide a secure and versatile unit of processing that the CLR can use to provide isolation between applications running in the same process."
	descWinDotNetLoadingCurrentAssemblies         = "This counter displays the current number of Assemblies loaded across all AppDomains in this application. If the Assembly is loaded as domain-neutral from multiple AppDomains then this counter is incremented once only. Assemblies can be loaded as domain-neutral when their code can be shared by all AppDomains or they can be loaded as domain-specific when their code is private to the AppDomain."
	descWinDotNetLoadingCurrentClassesLoaded      = "This counter displays the current number of classes loaded in all Assemblies."
	descWinDotNetLoadingTotalAppdomains           = "This counter displays the peak number of AppDomains loaded since the start of this application."
	descWinDotNetLoadingTotalappdomainsunloaded   = "This counter displays the total number of AppDomains unloaded since the start of the application. If an AppDomain is loaded and unloaded multiple times this counter would count each of those unloads as separate."
	descWinDotNetLoadingTotalAssemblies           = "This counter displays the total number of Assemblies loaded since the start of this application. If the Assembly is loaded as domain-neutral from multiple AppDomains then this counter is incremented once only."
	descWinDotNetLoadingTotalClassesLoaded        = "This counter displays the cumulative number of classes loaded in all Assemblies since the start of this application."
	descWinDotNetLoadingTotalNumberofLoadFailures = "This counter displays the peak number of classes that have failed to load since the start of the application. These load failures could be due to many reasons like inadequate security or illegal format."
)

type Win32_PerfRawData_NETFramework_NETCLRLoading struct {
	Currentappdomains         uint32
	CurrentAssemblies         uint32
	CurrentClassesLoaded      uint32
	Name                      string
	TotalAppdomains           uint32
	Totalappdomainsunloaded   uint32
	TotalAssemblies           uint32
	TotalClassesLoaded        uint32
	TotalNumberofLoadFailures uint32
}
