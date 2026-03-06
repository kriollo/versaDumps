export namespace main {
	
	export class LogFolder {
	    path: string;
	    extensions: string[];
	    filters: string[];
	    enabled: boolean;
	    format?: string;
	
	    static createFrom(source: any = {}) {
	        return new LogFolder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.extensions = source["extensions"];
	        this.filters = source["filters"];
	        this.enabled = source["enabled"];
	        this.format = source["format"];
	    }
	}
	export class Profile {
	    name: string;
	    server: string;
	    port: number;
	    theme?: string;
	    language?: string;
	    show_types?: boolean;
	    log_folders?: LogFolder[];
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.server = source["server"];
	        this.port = source["port"];
	        this.theme = source["theme"];
	        this.language = source["language"];
	        this.show_types = source["show_types"];
	        this.log_folders = this.convertValues(source["log_folders"], LogFolder);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
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
	export class WindowPosition {
	    x: number;
	    y: number;
	    width: number;
	    height: number;
	
	    static createFrom(source: any = {}) {
	        return new WindowPosition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.width = source["width"];
	        this.height = source["height"];
	    }
	}

}

