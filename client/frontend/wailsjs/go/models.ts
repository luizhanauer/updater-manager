export namespace main {
	
	export class UIApp {
	    id: string;
	    name: string;
	    description: string;
	    category: string;
	    icon_url: string;
	    status: string;
	    message: string;
	    is_installed: boolean;
	    auto_update: boolean;
	    progress: number;
	    downloaded_size: string;
	    total_size: string;
	    local_version: string;
	    remote_version: string;
	
	    static createFrom(source: any = {}) {
	        return new UIApp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.icon_url = source["icon_url"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.is_installed = source["is_installed"];
	        this.auto_update = source["auto_update"];
	        this.progress = source["progress"];
	        this.downloaded_size = source["downloaded_size"];
	        this.total_size = source["total_size"];
	        this.local_version = source["local_version"];
	        this.remote_version = source["remote_version"];
	    }
	}

}

