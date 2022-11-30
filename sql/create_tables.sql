CREATE DATABASE michman;

CREATE TABLE `project` (
	`ID` varchar(255),
	`Name` varchar(255) NOT NULL UNIQUE,
	`DisplayName` varchar(255) NOT NULL,
	`GroupID` varchar(255),
	`Description` TEXT,
	`DefaultImage` varchar(255) NOT NULL,
	`DefaultMasterFlavor` varchar(255) NOT NULL,
	`DefaultSlavesFlavor` varchar(255) NOT NULL,
	`DefaultStorageFlavor` varchar(255) NOT NULL,
	`DefaultMonitoringFlavor` varchar(255),
	PRIMARY KEY (`ID`)
);

CREATE TABLE `cluster` (
	`ID` varchar(255),
	`Name` varchar(255) NOT NULL UNIQUE,
	`DisplayName` varchar(255) NOT NULL,
	`HostURL` varchar(255), 
	`EntityStatus` varchar(32) NOT NULL,
	`ClusterType` varchar(255) NOT NULL, 
	`NSlaves` int NOT NULL,
	`MasterIP` varchar(255),
	`ProjectID` varchar(255) NOT NULL,
	`Description` TEXT,
	`Image` varchar(255) NOT NULL,
	`SSH_Keys` json,
	`Monitoring` boolean NOT NULL, 
	`MasterFlavor` varchar(255), 
	`SlavesFlavor` varchar(255), 
	`StorageFlavor` varchar(255), 
	`MonitoringFlavor` varchar(255),
	PRIMARY KEY (`ID`)
);


CREATE TABLE `service` (
	`ID` varchar(255),
	`Name` varchar(255) NOT NULL,
	`Type` varchar(255) NOT NULL,
	`ClusterRef` varchar(255) NOT NULL,
	`Config` TEXT,
	`DisplayName` varchar(255), 
	`EntityStatus` varchar(32),
	`Version` varchar(255) NOT NULL,
	`URL` varchar(255),
	`Description` TEXT,
	PRIMARY KEY (`ID`)
);

CREATE TABLE `template` (
	`ID` varchar(255) NOT NULL,
	`ProjectID` varchar(255),
	`Name` varchar(255) NOT NULL UNIQUE,
	`DisplayName` varchar(255) NOT NULL, 
	`NSlaves` int,
	`Description` TEXT,
	PRIMARY KEY (`ID`)
);

CREATE TABLE `health_configs` (
	`ID` varchar(255),
	`ParameterName` varchar(255) NOT NULL,
	`Description` varchar(255) ,
	`Type` varchar(255) NOT NULL,
	`DefaultValue` varchar(255) NOT NULL,
	`Required` boolean NOT NULL, 
	`AnsibleVarName` varchar(255) NOT NULL,
	`IsList` boolean NOT NULL, 
	`CheckType` varchar(255) NOT NULL,
	PRIMARY KEY (`ID`)
);

CREATE TABLE `health_check`(
	`ID` varchar(255) NOT NULL, 
	`CheckType` varchar(255) NOT NULL,
	`ServiceTypeID` varchar(255) NOT NULL UNIQUE,
	PRIMARY KEY (`ID`)
);

CREATE TABLE `service_type` (
	`ID` varchar(255) NOT NULL,
	`Type` varchar(255) NOT NULL UNIQUE,
	`Description` TEXT,
	`DefaultVersion` varchar(255) NOT NULL,
	`Class` varchar(32) NOT NULL,
	`AccessPort` varchar(32),
	PRIMARY KEY (`ID`)
);

CREATE TABLE `service_version` (
	`ID` varchar(255) ,
	`Version` varchar(255) NOT NULL,
	`Description` TEXT,
	`DownloadURL` TEXT,
	`ServiceTypeID` varchar(255) NOT NULL,
	PRIMARY KEY (`ID`)
);

CREATE TABLE `service_config` (
	`ID` varchar(255),
	`ParameterName` varchar(255) NOT NULL,
	`Type` varchar(32) NOT NULL,
	`PossibleValues` TEXT,
	`DefaultValue` varchar(255) NOT NULL,
	`Required` boolean NOT NULL,
	`Description` TEXT,
	`AnsibleVarName` varchar(255) NOT NULL,
	`IsList` boolean NOT NULL,
	`VersionID` varchar(255) NOT NULL,
	PRIMARY KEY (`ID`)
);

CREATE TABLE `service_dependency` (
	`ID` varchar(255) NOT NULL,
	`ServiceType` varchar(255) NOT NULL,
	`DefaultServiceVersion` varchar(255) NOT NULL,
	`Description` TEXT,
	`ServiceVersionID` varchar(255) NOT NULL,
	PRIMARY KEY (`ID`)
);

CREATE TABLE `image` (
	`ID` varchar(255) NOT NULL,
	`Name` varchar(255) NOT NULL UNIQUE,
	`AnsibleUser` varchar(255) NOT NULL,
	`CloudImageId` varchar(255) NOT NULL,
	PRIMARY KEY (`ID`)
);

CREATE TABLE `service_port` (
	`ID` varchar(255),
	`Port` varchar(32) NOT NULL UNIQUE,
	`ServiceTypeID` varchar(255) NOT NULL,
	`Description` TEXT,
	PRIMARY KEY (`ID`)
);

CREATE TABLE `flavor`(
	`ID` varchar(255), 
	`Name` varchar(255) NOT NULL UNIQUE,
	`VCPUs` int UNSIGNED NOT NULL, 
	`RAM` int UNSIGNED NOT NULL,
	`Disk` int UNSIGNED NOT NULL,
	PRIMARY KEY (`ID`)
);

CREATE TABLE `dependency_to_version` (
	`ServiceDependencyID` varchar(255) NOT NULL,
	`DependentVersionID` varchar(255) NOT NULL,
	PRIMARY KEY (`ServiceDependencyID`, `DependentVersionID`)
);

ALTER TABLE `project` ADD CONSTRAINT `Project_fk0` FOREIGN KEY (`DefaultImage`) REFERENCES `image`(`Name`);

ALTER TABLE `project` ADD CONSTRAINT `Project_fk1` FOREIGN KEY (`DefaultMasterFlavor`) REFERENCES `flavor`(`Name`);

ALTER TABLE `project` ADD CONSTRAINT `Project_fk2` FOREIGN KEY (`DefaultSlavesFlavor`) REFERENCES `flavor`(`Name`);

ALTER TABLE `project` ADD CONSTRAINT `Project_fk3` FOREIGN KEY (`DefaultStorageFlavor`) REFERENCES `flavor`(`Name`);

ALTER TABLE `project` ADD CONSTRAINT `Project_fk4` FOREIGN KEY (`DefaultMonitoringFlavor`) REFERENCES `flavor`(`Name`);

ALTER TABLE `cluster` ADD CONSTRAINT `Cluster_fk0` FOREIGN KEY (`ProjectID`) REFERENCES `project`(`ID`);

ALTER TABLE `cluster` ADD CONSTRAINT `Cluster_fk1` FOREIGN KEY (`Image`) REFERENCES `image`(`Name`);

ALTER TABLE `cluster` ADD CONSTRAINT `Cluster_fk2` FOREIGN KEY (`MasterFlavor`) REFERENCES `flavor`(`Name`);

ALTER TABLE `cluster` ADD CONSTRAINT `Cluster_fk3` FOREIGN KEY (`SlavesFlavor`) REFERENCES `flavor`(`Name`);

ALTER TABLE `cluster` ADD CONSTRAINT `Cluster_fk4` FOREIGN KEY (`StorageFlavor`) REFERENCES `flavor`(`Name`);

ALTER TABLE `cluster` ADD CONSTRAINT `Cluster_fk5` FOREIGN KEY (`MonitoringFlavor`) REFERENCES `flavor`(`Name`);

ALTER TABLE `service` ADD CONSTRAINT `Service_fk0` FOREIGN KEY (`Type`) REFERENCES `service_type`(`Type`);

ALTER TABLE `service` ADD CONSTRAINT `Service_fk1` FOREIGN KEY (`ClusterRef`) REFERENCES `cluster`(`ID`) ON DELETE CASCADE;

ALTER TABLE `template` ADD CONSTRAINT `Template_fk0` FOREIGN KEY (`ProjectID`) REFERENCES `project`(`ID`);

ALTER TABLE `service_version` ADD CONSTRAINT `ServiceVersion_fk0` FOREIGN KEY (`ServiceTypeID`) REFERENCES `service_type`(`ID`) ON DELETE CASCADE;

ALTER TABLE `service_config` ADD CONSTRAINT `ServiceConfig_fk0` FOREIGN KEY (`VersionID`) REFERENCES `service_version`(`ID`) ON DELETE CASCADE;

ALTER TABLE `service_dependency` ADD CONSTRAINT `ServiceDependency_fk0` FOREIGN KEY (`ServiceType`) REFERENCES `service_type`(`Type`);

ALTER TABLE `service_dependency` ADD CONSTRAINT `ServiceDependency_fk1` FOREIGN KEY (`ServiceVersionID`) REFERENCES `service_version`(`ID`) ON DELETE CASCADE;

ALTER TABLE `dependency_to_version` ADD CONSTRAINT `DependencyToVersion_fk0` FOREIGN KEY (`ServiceDependencyID`) REFERENCES `service_dependency`(`ID`) ON DELETE CASCADE;

ALTER TABLE `dependency_to_version` ADD CONSTRAINT `DependencyToVersion_fk1` FOREIGN KEY (`DependentVersionID`) REFERENCES `service_version`(`ID`);

ALTER TABLE `service_port` ADD CONSTRAINT `ServicePort_fk0` FOREIGN KEY (`ServiceTypeID`) REFERENCES `service_type`(`ID`) ON DELETE CASCADE;

ALTER TABLE `health_check` ADD CONSTRAINT `HealthCheck_fk0` FOREIGN KEY (`ServiceTypeID`) REFERENCES `service_type`(`ID`) ON DELETE CASCADE;

ALTER TABLE `health_configs` ADD CONSTRAINT `HealthConfig_fk0` FOREIGN KEY (`CheckType`) REFERENCES `health_check`(`ID`) ON DELETE CASCADE;

