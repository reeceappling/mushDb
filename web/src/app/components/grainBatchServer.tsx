import {Note} from "@/app/components/formSubcomponents/notes";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {GrainBatchSelector, NewGrainBatchForm} from "@/app/components/grainBatchClient";
import {ACL, TestAcl} from "@/app/components/accessControlServer";

export function TestGrainBatchOkFull() {
    return new GrainBatchData({
        _id: "(GRAIN BATCH ID HERE)",
        soakTimeHrs: 9,
        boilTimeMins: 30,
        dryTimeHours: 4,
        recipe: ("GRAIN RECIPE ID HERE"),
        creationDate: Date.now(),
        notes: [{time: Date.now(), note: "(TEST NOTE 1)"}, {time: Date.now() + 2000, note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
        acl: TestAcl(),
    })
}

export interface GrainBatchData {
    _id: string
    creationDate: number
    recipe: string
    soakTimeHrs?: number
    boilTimeMins?: number
    dryTimeHours?: number
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class GrainBatchData {
    // Accept a single object containing the fields
    constructor(init?: Partial<GrainBatchData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "grainBatch"
    }
}

export function GrainBatchSelectorCloseable(sp: SelectorProps<GrainBatchData>) { // TODO: use
    const doSel = (val?: GrainBatchData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<GrainBatchData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Grain Batch List",
        createTxt: "Create Grain Batch",
        lowercase: "grain batch",
        creatorInPage: sp.creatorInPage,
        createEndpt: "grainBatch",
        createSelector:(selHdl: (onSelect: GrainBatchData) => void)=>{
            return <GrainBatchSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: GrainBatchData) => void)=>{
            return <NewGrainBatchForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}
