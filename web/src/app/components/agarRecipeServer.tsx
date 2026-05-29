import {Note} from "@/app/components/formSubcomponents/notes";
import {Liquid} from "@/app/components/formSubcomponents/liquids";
import {Nutrient} from "@/app/components/formSubcomponents/nutrients";
import {Sugar} from "@/app/components/formSubcomponents/sugars";
import {Antibiotic} from "@/app/components/formSubcomponents/antibiotic";
import {Additive} from "@/app/components/formSubcomponents/additives";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {AgarRecipeSelector, NewAgarRecipeForm} from "@/app/components/agarRecipeClient";

export function TestAgarRecipeOk(){
    const a: AgarRecipeData = {
        _id: "(AGAR RECIPE ID HERE)",
        name: "(AGAR RECIPE NAME HERE)",
        liquids: [{name:"Water",pct:100}],
        agar: 20, // g/L
        standard: true,
        nutrients: [{nutrient:"Oats",amount:17.2,unit:"handfuls"}],
        sugars: [{type:"Glucose",amount:100,unit:"tons"}],
        additives: [],// TODO: this
        antibiotics: [], // TODO: this
        notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
    }
    return a
}

export interface AgarRecipeData {
    _id: string
    name: string
    liquids: Liquid[]
    agar: number
    standard: boolean
    nutrients?: Nutrient[]
    sugars?: Sugar[]
    additives?: Additive[]
    antibiotics?: Antibiotic[]
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}

export function AgarRecipeSelectorCloseable(sp: SelectorProps<AgarRecipeData>) {
    const doSel = (val?: AgarRecipeData): void => {
        val && sp.doSelect(val)
    }
    return <CloseableSelector<AgarRecipeData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        msgTxt: "ChannelTextNewAgarRecipe",
        closeTxt: "Close Recipe List",
        createTxt: "Create Agar Recipe",
        lowercase: "agar recipe",
        creatorInPage: sp.creatorInPage,
        createEndpt: "agarRecipe",
        getId: (v: AgarRecipeData) => v._id,
        createSelector: (selHdl: (onSelect: AgarRecipeData) => void) => {
            return <AgarRecipeSelector allowCreate={sp.allowCreation} doSelect={(v) => {
                v && selHdl(v)
            }}/>
        },
        createCreator: (selHdl: (onSelect: AgarRecipeData) => void) => {
            return <NewAgarRecipeForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}
