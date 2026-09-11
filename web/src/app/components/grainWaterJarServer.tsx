import {Note} from "@/app/components/formSubcomponents/notes";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {GrainBatchSelector, NewGrainBatchForm} from "@/app/components/grainBatchClient";
import {ACL} from "@/app/components/accessControlServer";
import {GrainWaterJarSelector, NewGrainWaterJarForm} from "@/app/components/grainWaterJarClient";

// export function TestGrainBatchOkFull() {
//     return new GrainBatchData({
//         _id: "(GRAIN BATCH ID HERE)",
//         soakTimeHrs: 9,
//         boilTimeMins: 30,
//         dryTimeHours: 4,
//         recipe: ("GRAIN RECIPE ID HERE"),
//         creationDate: Date.now(),
//         notes: [{time: Date.now(), note: "(TEST NOTE 1)"}, {time: Date.now() + 2000, note: "(TEST NOTE 2)"}],
//         lastUpdated: 789,
//         acl: TestAcl(),
//     })
// }

export interface GrainWaterJarData {
    _id: string
    grainBatch: string
    creationDate: number
    notes?: Note[]
    lastUpdated: number
    acl: ACL
    disposed?: number
}
export class GrainWaterJarData {
    // Accept a single object containing the fields
    constructor(init?: Partial<GrainWaterJarData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "grainWaterJar" // TODO: ensure ok
    }
    public description(): string {
        return `Grain Water Jar ${this._id}. From batch ${this.grainBatch}. Created on ${new Date(this.creationDate).toISOString()}. Last updated on ${new Date(this.lastUpdated).toISOString()}`
    }
}

export function GrainWaterJarSelectorCloseable(sp: SelectorProps<GrainWaterJarData>) {
    const doSel = (val?: GrainWaterJarData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<GrainWaterJarData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Grain Water Jar List",
        createTxt: "Create Grain Water Jar",
        lowercase: "grainwater jar",
        creatorInPage: sp.creatorInPage,
        createEndpt: "grainWaterJar", // TODO: ensure ok
        createSelector:(selHdl: (onSelect: GrainWaterJarData) => void)=>{
            return <GrainWaterJarSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: GrainWaterJarData) => void)=>{
            return <NewGrainWaterJarForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}
