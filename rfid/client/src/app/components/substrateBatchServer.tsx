import {Note} from "@/app/components/formSubcomponents/notes";
import {ACL} from "@/app/components/accessControlServer";

export function TestSubstrateBatchOkStd(std: boolean){
    let a: SubstrateBatchData = TestSubstrateBatchOk()
    return a
}
export function TestSubstrateBatchOk(){
    const a: SubstrateBatchData = {
        _id: "(SUBSTR BATCH ID HERE)",
        creationDate: 567,
        recipe: "(RECIPE ID HERE)",
        notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
    }
    return a
}

export interface SubstrateBatchData {
    _id: string
    creationDate: number
    recipe: string
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}