import {ACL, TestAcl} from "@/app/components/accessControlServer";
import {Note} from "@/app/components/formSubcomponents/notes";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {NewSubstrateRecipeForm, SubstrateRecipeSelector} from "@/app/components/substrateRecipeClient";
import {CapitalizeFirstLetter} from "@/app/components/commonServer";

export function TestSubstrateRecipeOkStd(std: boolean){
    const a: SubstrateRecipeData = TestSubstrateRecipeOk()
    a.standard =std
    return a
}
export function TestSubstrateRecipeOk(){
    return new SubstrateRecipeData({
        _id: "(SUBSTR ID HERE)",
        name: "(SUBSTR NAME HERE)",
        standard: false,
        aliases: ["(Alias 1)","(Alias 2)"],
        notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
        acl: TestAcl(), // TODO: do we want this?
    })
}

export interface SubstrateRecipeData {
    _id: string
    name: string
    standard: boolean
    aliases?: string[]
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class SubstrateRecipeData {
    // Accept a single object containing the fields
    constructor(init?: Partial<SubstrateRecipeData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public getIdUrlEncoded(): string {
        return encodeURI(this.getId())
    }
    public entryType(): string {
        return "substrateRecipe"
    }
    public description(): string {
        const standardPart = this.standard?`standard `:''
        const firstSentence = CapitalizeFirstLetter(`${standardPart}substrate recipe ${this.name}`)
        const lastSent = `Last updated on ${new Date(this.lastUpdated).toISOString()}`
        return `${firstSentence}. ${lastSent}`
    }
}

// TODO: VALIDATE WORKS!
export function SubstrateRecipeSelectorCloseable(sp: SelectorProps<SubstrateRecipeData>){ // TODO: likely overhaul
    return <CloseableSelector<SubstrateRecipeData> props={{
        allowCreation: sp.allowCreation,
        doSelect: sp.doSelect, // For selecting normally
        closeTxt: "Close Substrate Recipe List",
        createTxt: "Create Substrate Recipe",
        createEndpt: "substrateRecipe",
        lowercase: "substrate recipe",
        creatorInPage: sp.creatorInPage,
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