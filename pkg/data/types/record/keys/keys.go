package keys

// Record - the unique key type for data.Record entries
type Record = string

const (
	// Key of "DESCRIPTOR" type Records
	Descriptor Record = "DESCRIPTOR"

	// Key of "ARTIFACT_REMOVE" type Records
	Del = "del"

	// Key of repository ID entry, that contains {@link String}.
	RepositoryId = "repositoryId"

	// Key of all groups list entry, that contains {@link java.util.List<String>}.
	AllGroups     = "allGroups"
	AllGroupsList = "allGroupsList"

	/**
	 * Key of root groups list entry, that contains {@link java.util.List<String>}.
	 */
	RootGroups     = "rootGroups"
	RootGroupsList = "rootGroupsList"

	/**
	 * Key of index record modification (added to index or removed from index) timestamp entry, that contains {@link
	 * Long}.
	 */
	RecordModified = "recordModified"

	/**
	 * Key of artifact groupId entry, that contains {@link String}.
	 */
	GroupID = "groupId"

	/**
	 * Key of artifact artifactId entry, that contains {@link String}.
	 */
	ArtifactID = "artifactId"

	/**
	 * Key of artifact version entry, that contains {@link String}.
	 */
	Version = "version"

	/**
	 * Key of artifact classifier entry, that contains {@link String}.
	 */
	Classifier = "classifier"

	/**
	 * Key of artifact packaging entry, that contains {@link String}.
	 */
	Packaging = "packaging"

	/**
	 * Key of artifact file extension, that contains {@link String}.
	 */
	FileExtension = "fileExtension"

	/**
	 * Key of artifact file last modified timestamp, that contains {@link Long}.
	 */
	FileModified = "fileModified"

	/**
	 * Key of artifact file size in bytes, that contains {@link Long}.
	 */
	FileSize = "fileSize"

	/**
	 * Key of artifact Sources presence flag, that contains {@link Boolean}.
	 */
	HasSources = "hasSources"

	/**
	 * Key of artifact Javadoc presence flag, that contains {@link Boolean}.
	 */
	HasJavadoc = "hasJavadoc"

	/**
	 * Key of artifact signature presence flag, that contains {@link Boolean}.
	 */
	HasSignature = "hasSignature"

	/**
	 * Key of artifact name (as set in POM), that contains {@link String}.
	 */
	Name = "name"

	/**
	 * Key of artifact description (as set in POM), that contains {@link String}.
	 */
	Description = "description"

	/**
	 * Key of artifact SHA1 digest, that contains {@link String}.
	 */
	SHA1 = "sha1"

	/**
	 * Key of artifact contained class names, that contains {@link java.util.List<String>}. Extracted by {@code
	 * JarFileContentsIndexCreator}.
	 */
	Classnames = "classNames"

	/**
	 * Key of plugin artifact prefix, that contains {@link String}. Extracted by {@code
	 * MavenPluginArtifactInfoIndexCreator}.
	 */
	PluginPrefix = "pluginPrefix"

	/**
	 * Key of plugin artifact goals, that contains {@link java.util.List<String>}. Extracted by {@code
	 * MavenPluginArtifactInfoIndexCreator}.
	 */
	PluginGoals = "pluginGoals"

	/**
	 * Key of OSGi "Bundle-SymbolicName" manifest entry, that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	OSGIBundleSymbolicName = "Bundle-SymbolicName"

	/**
	 * Key of OSGi "Bundle-Version" manifest entry, that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	OSGIBundleVersion = "Bundle-Version"

	/**
	 * Key of OSGi "Export-Package" manifest entry, that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	OSGIExportPackage = "Export-Package"

	/**
	 * Key of OSGi "Export-Service" manifest entry, that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	OSGIExportService = "Export-Service"

	/**
	 * Key of OSGi "Bundle-Description" manifest entry, that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	OSGIBundleDescription = "Bundle-Description"

	/**
	 * Key of OSGi "Bundle-Name" manifest entry, that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	OSGIBundleName = "Bundle-Name"

	/**
	 * Key of OSGi "Bundle-License" manifest entry, that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	OSGIBundleLicense = "Bundle-License"

	/**
	 * Key of OSGi "Bundle-DocURL" manifest entry, that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	OSGIExportDocURL = "Bundle-DocURL"

	/**
	 * Key of OSGi "Import-Package" manifest entry, that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	OSGIImportPackage = "Import-Package"

	/**
	 * Key of OSGi "Require-Bundle" manifest entry, that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	OSGIRequireBundle = "Require-Bundle"

	/**
	 * Key of OSGi "Provide-Capability" manifest entry, that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	OSGIProvideCapability = "Provide-Capability"

	/**
	 * Key of OSGi "Require-Capability" manifest entry, that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	OSGIRequireCapability = "Require-Capability"

	/**
	 * Key of OSGi "Fragment-Host" manifest entry, that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	OSGIFragmentHost = "Fragment-Host"

	/**
	 * Key of deprecated OSGi "Bundle-RequiredExecutionEnvironment" manifest entry, that contains {@link String}.
	 * Extracted by {@code OsgiArtifactIndexCreator}.
	 */
	OSGIBREE = "Bundle-RequiredExecutionEnvironment"

	/**
	 * Key for SHA-256 checksum  needed for OSGI content capability that contains {@link String}. Extracted by {@code
	 * OsgiArtifactIndexCreator}.
	 */
	SHA256 = "sha256"
)
