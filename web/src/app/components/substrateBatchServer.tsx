import {Note} from "@/app/components/formSubcomponents/notes";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {NewSubstrateBatchForm, SubstrateBatchSelector} from "@/app/components/substrateBatchClient";

export function TestSubstrateBatchOkStd(std: boolean){
    let a: SubstrateBatchData = TestSubstrateBatchOk()
    return a
}
export function TestSubstrateBatchOk(){
    return new SubstrateBatchData({
        _id: "(SUBSTR BATCH ID HERE)",
        creationDate: 567,
        recipe: "(RECIPE ID HERE)",
        notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
    })
}

export interface SubstrateBatchData {
    _id: string
    creationDate: number
    recipe: string
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}
export class SubstrateBatchData {
    // Accept a single object containing the fields
    constructor(init?: Partial<SubstrateBatchData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "substrateBatch"
    }
}

// TODO: VALIDATE WORKS!
export function SubstrateBatchSelectorCloseable(sp: SelectorProps<SubstrateBatchData>){ // TODO: likely overhaul
    return <CloseableSelector<SubstrateBatchData> props={{
        allowCreation: sp.allowCreation,
        doSelect: sp.doSelect, // For selecting normally
        closeTxt: "Close Substrate Batch List",
        createTxt: "Create Substrate Batch",
        createEndpt: "substrateBatch",
        lowercase: "substrate batch",
        creatorInPage: sp.creatorInPage,
        createSelector:(selHdl: (onSelect: SubstrateBatchData) => void)=>{
            return <SubstrateBatchSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: SubstrateBatchData) => void)=>{
            return <NewSubstrateBatchForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}