import {ACL} from "@/app/components/accessControlServer";
import {Note} from "@/app/components/formSubcomponents/notes";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {NewSubstrateRecipeForm, SubstrateRecipeSelector} from "@/app/components/substrateRecipeClient";

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

// TODO: VALIDATE WORKS!
export function SubstrateRecipeSelectorCloseable(sp: SelectorProps<SubstrateRecipeData>){ // TODO: likely overhaul
    return <CloseableSelector<SubstrateRecipeData> props={{
        allowCreation: sp.allowCreation,
        doSelect: sp.doSelect, // For selecting normally
        msgTxt: "", // TODO: del?
        closeTxt: "Close Substrate Recipe List",
        createTxt: "Create Substrate Recipe",
        createEndpt: "substrateRecipe",
        lowercase: "substrate recipe",
        creatorInPage: sp.creatorInPage,
        getId: (v: SubstrateRecipeData)=>v._id,
        createSelector:(selHdl: (onSelect: SubstrateRecipeData) => void)=>{
            return <SubstrateRecipeSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: SubstrateRecipeData) => void)=>{
            return <NewSubstrateRecipeForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}