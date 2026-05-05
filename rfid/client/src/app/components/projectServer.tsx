import {Note} from "@/app/components/formSubcomponents/notes";
import {SelectorProps} from "@/app/components/selector";
import {ProjectPerms} from "@/app/components/perms";

export function TestProjectOk(){
    let perms = new Map<string, string>();
    perms.set("USERNAME 1", "admin")
    perms.set("USERNAME 2", "write")
    perms.set("USERNAME 3", "read")
    const a: ProjectData = {
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
        perms: perms,
    }
    return a
}

export type ProjectData = {
    _id: string // project name
    creationDate: number
    completed?: number
    notes?: Note[]
    lastUpdated: number
    perms?: Map<string, string> // Map of userId to canWrite // TODO: consider changing from bool to "read/write/admin" if serialization of mapped undefineds gets weird...
}

// Confirmed to be working without going to get data
export function ProjectSelector(sp: SelectorProps<ProjectData>){
    // TODO: FIX THIS?
    return <select className={"tailwindSelector"} onChange={e => { // TODO: DISABLE THIS RETURN!
        // TODO: FIX NEXT LINE
        sp.doSelect({_id: e.currentTarget.value, creationDate: 0, lastUpdated: 0, perms: new Map<string, string>()})
    }}>
        <option value={"A"}>{"A"}</option>
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

export const ChannelTextNewProject = "newProject"