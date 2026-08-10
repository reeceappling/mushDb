import {Note} from "@/app/components/formSubcomponents/notes";


import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {NewWaterJarForm, WaterJarSelector} from "@/app/components/waterJarClient";

// export function TestWaterOk(){
//     const now = new Date().getTime()
//     const testNote = ()=>{
//         return {time: new Date().getTime(), note:"TEST_NOTE_TEXT_HERE"}
//     }
//     const testNotes: Note[] = [testNote(), testNote(), testNote()]
//     const a: WaterJarData = {
//         _id: "(WATER JAR ID HERE)",
//         creationDate: now,
//         pcRun: "(PC RUN ID HERE)",
//         notes: [...testNotes],
//         lastUpdated: 789,
//         acl: TestAcl(), // TODO: do we want this?
//     }
//     return a
// }
export interface WaterJarData {
    _id: string
    creationDate: number
    pcRun: string
    notes?: Note[]
    disposed?: number
    lastUpdated: number
    acl: ACL
}
export class WaterJarData {
    // Accept a single object containing the fields
    constructor(init?: Partial<WaterJarData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "waterJar"
    }
    public description(): string {
        if(this.disposed !== undefined) {
            return `Water jar ${this._id}. Created on ${new Date(this.creationDate).toISOString()}. Disposed on ${new Date(this.disposed).toISOString()}`
        }
        return `Water jar ${this._id}. Last updated on ${new Date(this.lastUpdated).toISOString()}. Created on ${new Date(this.creationDate).toISOString()}. Pc Run ${this.pcRun}`
    }
}

export function WaterJarSelectorCloseable(sp: SelectorProps<WaterJarData>) {
    const doSel = (val?: WaterJarData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<WaterJarData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Water Jar List",
        createTxt: "Create Water Jar",
        lowercase: "water jar",
        creatorInPage: sp.creatorInPage,
        createEndpt: "waterJar",
        createSelector:(selHdl: (onSelect: WaterJarData) => void)=>{
            return <WaterJarSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: WaterJarData) => void)=>{
            return <NewWaterJarForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}