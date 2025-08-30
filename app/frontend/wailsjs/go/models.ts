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

}

