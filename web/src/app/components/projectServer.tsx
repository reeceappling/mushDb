import {Note} from "@/app/components/formSubcomponents/notes";
import {SelectorProps} from "@/app/components/selector";

export function TestProjectOk(){
    const perms = new Map<string, string>();
    perms.set("USERNAME 1", "admin")
    perms.set("USERNAME 2", "write")
    perms.set("USERNAME 3", "read")
    return new ProjectData({
        _id: "(PROJECT NAME HERE)",
        creationDate: 123,
        completed: 456, // Optional
        notes: [{
            time: 123,
            note: "(NOTE 1)"
        },{
            time: 456,
            note: "(NOTE 2)"
        }],
        lastUpdated: 789,
        perms: perms, // TODO: is this ok?
    })
}
export function TestProjectOk2(){
    const perms = new Map<string, string>();
    perms.set("USERNAME 1", "admin")
    perms.set("USERNAME 2", "write")
    perms.set("USERNAME 3", "read")
    return new ProjectData({
        _id: "(PROJECT NAME HERE)",
        creationDate: 123,
        notes: [{
            time: 123,
            note: "(NOTE 1)"
        },{
            time: 456,
            note: "(NOTE 2)"
        }],
        lastUpdated: 789,
        perms: perms,
    })
}

export interface ProjectData {
    _id: string, // project name
    creationDate: number,
    completed?: number,
    notes?: Note[],
    lastUpdated: number,
    perms?: Map<string, string>, // Map of userId to "read/write/admin"
}
export class ProjectData {
    // Accept a single object containing the fields
    constructor(init?: Partial<ProjectData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public getIdUrlEncoded(): string {
        return encodeURI(this._id)
    }
    public entryType(): string {
        return "project"
    }
}

// Confirmed to be working without going to get data // TODO: delete if unused? Make closeable?
export function ProjectSelector(sp: SelectorProps<ProjectData>){
    // TODO: FIX THIS?
    return <select className={"tailwindSelector"} onChange={e => { // TODO: DISABLE THIS RETURN!
        e.stopPropagation() // TODO: ok?
        sp.doSelect(new ProjectData({_id: e.currentTarget.value, creationDate: 0, lastUpdated: 0, perms: new Map<string, string>()}))
    }}>
        <option value={"A"}>{"A"}</option>{/* TODO: gross and change to realistic*/}
        <option value={"B"}>{"B"}</option>
        <option value={"C"}>{"C"}</option>
    </select>
    // return <RecentSelector props={{ // TODO: REENABLE ME!
    //     msgTxt: ChannelTextNewProject,
    //     recentEndpt: "projects",
    //     assertType: AssertProject,
    //     closeTxt: "Close Project List",
    //     createTxt: "Create Project",
    //     createEndpt: "project",
    //     lowercase: "project",
    //     inline: (inlineIn: InlineProps<ProjectData>)=>{return ProjectInline(inlineIn)},
    //     getId: (v: ProjectData)=>v._id,
    //     doSelect: sp.doSelect,
    //     allowCreation: sp.allowCreation,
    //     creatorInPage: sp.creatorInPage,
    // }}>
    //     <NewProjectForm onCreate={sp.doSelect}/>
    // </RecentSelector>
}