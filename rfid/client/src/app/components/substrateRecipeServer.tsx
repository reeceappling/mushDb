import {ACL} from "@/app/components/accessControlServer";

export function TestSubstrateRecipeOkStd(std: boolean){
    let a: SubstrateRecipeData = TestSubstrateRecipeOk()
    a.standard =std
    return a
}
export function TestSubstrateRecipeOk(){
    const a: SubstrateRecipeData = {
        _id: "(SUBSTR ID HERE)",
        name: "(SUBSTR NAME HERE)",
        standard: false,
        aliases: ["(Alias 1)","(Alias 2)"],
        notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
    }
    return a
}

export interface SubstrateRecipeData {
    _id: string
    name: string
    standard: boolean
    aliases?: string[]
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}