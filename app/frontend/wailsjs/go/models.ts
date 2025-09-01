export namespace main {
	
	export class Config {
	    Server: string;
	    Port: number;
	    Theme: string;
	    Lang: string;
	    ShowTypes: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Server = source["Server"];
	        this.Port = source["Port"];
	        this.Theme = source["Theme"];
	        this.Lang = source["Lang"];
	        this.ShowTypes = source["ShowTypes"];
	    }
	}
	export class UpdateInfo {
	    available: boolean;
	    version: string;
	    description: string;
	    downloadUrl: string;
	    releaseUrl: string;
	    size: number;
	    currentVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.version = source["version"];
	        this.description = source["description"];
	        this.downloadUrl = source["downloadUrl"];
	        this.releaseUrl = source["releaseUrl"];
	        this.size = source["size"];
	        this.currentVersion = source["currentVersion"];
	    }
	}

}

