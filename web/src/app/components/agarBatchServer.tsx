import {Note} from "@/app/components/formSubcomponents/notes";
import {
    AgarBatchSelector,
    NewAgarBatchForm
} from "@/app/components/agarBatchClient";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {ACL, TestAcl} from "@/app/components/accessControlServer";

export function TestAgarBatchOk() {
    return new AgarBatchData({
        _id: "(Batch ID HERE)",
        color: "Clear",
        pcRun: "(Run ID HERE)",
        agarRecipe: "(Recipe ID HERE)",
        notes: [{time: Date.now(), note: "(TEST NOTE 1)"}, {time: Date.now() + 2000, note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
        acl: TestAcl(),
    });
}

export interface AgarBatchData {
    _id: string
    color: string
    pcRun: string
    agarRecipe: string
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class AgarBatchData {
    // Accept a single object containing the fields
    constructor(init?: Partial<AgarBatchData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "agarBatch"
    }
}

export function AgarBatchSelectorCloseable(sp: SelectorProps<AgarBatchData>) {
    const doSel = (val?: AgarBatchData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<AgarBatchData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Batch List",
        createTxt: "Create Agar Batch",
        lowercase: "agar batch",
        creatorInPage: sp.creatorInPage,
        createEndpt: "agarBatch",
        createSelector:(selHdl: (onSelect: AgarBatchData) => void)=>{
            return <AgarBatchSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: AgarBatchData) => void)=>{
            return <NewAgarBatchForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}

export const ChannelTextNewAgarBatch = "newAgarBatch"