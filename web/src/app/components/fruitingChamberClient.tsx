'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {
    AllEntries,
    Data,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea, {NumbersOnlyFromText} from "@/app/components/formSubcomponents/date";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import {InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {
    dataFor,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateMultipartRequest,
    ErrHandler,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    FloatInput,
    ImportDisplayInput,
    ImportEntryFormWrapper,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    MainCollectionInputOrRead,
    MultipartImportRequest,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    RequiredKey,
    resolveContamsFormData,
    resolvePicsFormData,
    SelectorWrapper,
    setFormData,
    setFormImages,
} from "@/app/components/common";
import {
    ErrorDisplay,
    GensFormDisplay,
    MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay
} from "@/app/components/formSubcomponents/commonClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {SubstrateRecipeArea, SubstrateRecipeSelector} from "@/app/components/substrateRecipeClient";
import {
    ContaminationForm,
    ContamsDisplay,
    InitialContamState,
    IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SaleArea} from "@/app/components/saleClient";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {ExistingSpeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {SubstrateBatchArea, SubstrateBatchSelector} from "@/app/components/substrateBatchClient";
import {SelectorFor} from "@/app/components/selector";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {InputNumber} from "@/app/components/formSubcomponents/numericInput";
import {OnViewCreatorsQuadColArea, OvcForNewFruit} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";

export function AssertFruitingChamber(input: any): asserts input is FruitingChamberData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['recipe', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
        ['cupsGrain', 'number'],
        ['mixedSubstratePerGrain', 'number'],
        ['casingPerGrain', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('FruitingChamber assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['substrateBatch', 'string'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['knownFruitable', 'boolean'],
        ['sale', 'string'],
        ['disposed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('FruitingChamber assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    let complexRequiredKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl],
    ])
    for (let [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('FruitingChamber assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Fruiting chamber assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['flushes', IsValidPicWithNotesIncoming],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('FruitingChamber assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function FruitingChamberDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {
        AssertFruitingChamber(data)
        const [initial, setInitial] = useState(data)

        const existingNotes: Note[] = initial.notes || []
        const initNotes: Data<Note>[] = existingNotes.map((n) => {
            return {data: n, disabled: false}
        })
        // Basic objects
        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [sale, setSale] = useState(initial.sale)
        const [notes, setNotes] = useState<AllEntries<Note>>({existing: initNotes, new: []})
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const [writeTo, setWriteTo] = useState<string | undefined>()

        // Image-containing
        const [pics, setPics] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [flushes, setFlushes] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.flushes))

        // Helper states
        const [transfersOut, setTransfersOut] = useState(initial.transfersOut || [])
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: FruitingChamberData) => {
            setInitial(updated)
            setKnownFruitable(updated.knownFruitable)
            setDisposed(updated.disposed)
            setSale(updated.sale)
            setNotes({existing: dataFor(updated.notes), new: []})
            setAcl(updated.acl)
            setPics(InitialPicsEntries(updated.pics))
            setContams(InitialContamState(updated.contamination))
            setFlushes(InitialPicsEntries(updated.flushes))
            setTransfersOut(updated.transfersOut || [])
            setErr(undefined)
        }
        const cookies = useContext(CookiesContext)
        const fruitingChamberSubmit = () => {
            let formData = new FormData()
            let dataObj: any = {
                knownFruitable: knownFruitable,
                disposed: disposed,
                sale: sale,
                notes: notes,
                writeTagTo: writeTo,
                acl: MarshalAcl(acl),
            }
            try {
                // Pics
                let picsInfo = resolvePicsFormData(pics)
                let newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Contams
                let contamsInfo = resolveContamsFormData(contams)
                let newContams = contamsInfo.images
                dataObj.contams = contamsInfo.obj
                // Flushes
                let flushesInfo = resolvePicsFormData(flushes)
                let newFlushes = flushesInfo.images
                dataObj.flushes = flushesInfo.obj
                // Set data on form
                setFormData(formData, dataObj)
                setFormImages(formData, "newPic", newImages)
                setFormImages(formData, "newContam", newContams)
                setFormImages(formData, "newFlush", newFlushes)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            DoUpdateMultipartRequest("fruitingChamber",initial._id, formData, AssertFruitingChamber, allCookies(cookies))
                .then(v=>{
                    updateInitial(new FruitingChamberData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            OvcForNewFruit(data._id, "fruitingChamber", allCookies(cookies)),
            WriteRfidOvcArea(initial._id),
            // TODO: xfers? OvcForXfers(data._id, "fruit", ["plate","slant","jar","stasisTube"], "Clone/Transfer Fruit"), // TODO: ensure list correct //OVC for clone to plate? (transfer)
            // TODO: spore swab directly from box? (should also create a fruit in the interim)
            // TODO: spore print directly from box? (should also create a fruit in the interim)
        ]
        return (
            <DisplayFormWrapper entryType={"fruitingChamber"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID id={data._id} txt={"Fruiting Chamber"} entryType={"fruitingChamber"}/>
                <MostRecentImageDisplay data={initial.mostRecentImage}
                                        headerLevel={headerLevel}/>{/* Most recent image! */}
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                                    disposed={initial.disposed} readonly={readonly}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <SubstrateRecipeArea id={initial.recipe} headerLevel={headerLevel} readonly={true}/>
                        <SubstrateBatchArea id={initial.substrateBatch} readonly={true}/>
                        <div>{"Cups grain: " + data.cupsGrain}</div>
                        <div>{"Substrate mixed with grain: " + (data.mixedSubstratePerGrain * data.cupsGrain)}</div>
                        <div>{"Casing: " + (data.casingPerGrain * data.cupsGrain)}</div>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <GensFormDisplay gensSinceSpore={initial.genSpore}
                                         gensSinceFruitOrSpore={initial.genFruitOrSpore}
                                         headerLevel={headerLevel}/>{/* Gens since spore and spore/fruit*/}
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                        {/*<SpeciesSubspeciesFormArea species={initial.species} subspecies={initial.subspecies}/>*/}
                        <InnocDisplay innoc={initial.innoc} openInNewTab={false}/>{/* Innoc */}
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType}
                                       headerLevel={headerLevel}/>{/* Parent */}
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <KnownFruitableArea initial={knownFruitable} readonly={readonly} headerLevel={headerLevel}
                                            doSelect={setKnownFruitable}/>{/* KnownFruitable */}
                        <SaleArea sale={sale} setSale={setSale} readonly={readonly} headerLevel={headerLevel}
                                  canCreateSale={true}/>{/* Sale */}
                    </FlexedSinglesGroup>
                </FlexedArea>
                <TransfersOutDisplay thisId={initial._id} thisEntryType={"fruitingChamber"}
                                     transfersOut={initial.transfersOut} allowNewTransferCreation={false}/>

                <PicsDisplay pix={initial.pics || []} readonly={readonly} updateParent={setPics}/>{/* Pics */}
                {/* Flushes */}<PicsDisplay pix={initial.flushes || []} readonly={readonly}
                                            updateParent={setFlushes} addButtonText={"Create New Flush"}/>
                <ContamsDisplay initial={initial.contamination || []} updateParent={setContams}
                                readonly={readonly} headerLevel={headerLevel}/>{/* Contamination */}
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay initial={acl} readonly={readonly} updateParent={setAcl}/>
                </TogglableAreaWithDepth>
                {readonly ? null :
                    <button className={"bottomButton greenButton"} onClick={(e) => {
                        e.stopPropagation();
                        fruitingChamberSubmit()
                    }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Fruiting Chamber data format incorrect: " + err}</div>
    }
}

export function VolumeSelector({initialVal, initialUnit, updateNumberOfCups}: {
    initialVal?: number,
    initialUnit?: string,
    updateNumberOfCups: (n: number) => void
}) {
    // TODO: INITIAL NUMBERS AND UNITS
    const mulForUnit = (u: string) => {
        switch (u) {
            case "cups":
                return 1
            case "pints":
                return 2
            case "quarts":
                return 4
            default:
                return 0
        }
    }
    const [u, setUnit] = useState(initialUnit || "quarts")
    const [val, setVal] = useState<number>(initialVal || 1)
    const updateNumber = (s: string) => {
        let n = NumbersOnlyFromText(s)
        updateNumberOfCups(n * mulForUnit(u))
        setVal(n)
    }

    const changeUnit = (s: string) => {
        updateNumberOfCups(val * mulForUnit(s))
        setUnit(s)
    }

    updateNumberOfCups(val * mulForUnit(u))
    return <>
        <InputNumber min={1} max={10000} readonly={false} mode={"floating"} value={val.toString()} onChange={s => { // TODO: validate working properly
            if (s) {
                updateNumber(s)
            } else {
                updateNumber("0")
            }
        }} step={1}/>
        {/* Volume unit selector */}
        <SelectorFor options={["cups", "pints", "quarts"]} initial={u} updateParent={changeUnit} disabled={false}/>
    </>
}

export function NewFruitingChamberForm({handlers, substrateBatchIn, parent}: {
    handlers: NewEntryInput<FruitingChamberData>,
    substrateBatchIn?: SubstrateBatchData,
    parent?: string
}) {
    const [subBatch, setSubBatch] = useState<SubstrateBatchData | undefined>(substrateBatchIn)
    const [parentId, setParentId] = useState<string | undefined>() // TODO: do we even want this in the initial one? Keep it optional?
    const [volumeGrainCups, setVolumeGrainCups] = useState<number>(0)
    const [mixedSubCups, setMixedSubCups] = useState<number>(0)
    const [casingCups, setCasingCups] = useState<number>(0)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()

    const cookies = useContext(CookiesContext)
    const newFruitingChamberSubmit = () => {
        if (!subBatch) {
            setErr("substrate batch must be selected")
            return
        }
        if (volumeGrainCups <= 0) {
            setErr("casing must be a positive value")
            return
        }
        if (casingCups < 0) {
            setErr("casing must be a positive or zero value")
            return
        }
        if (mixedSubCups <= 0) {
            setErr("mixed substrate must be a positive number")
            return
        }
        let body: any = {
            substrateBatch: subBatch._id,
            parentJar: parentId,
            grainCups: volumeGrainCups,
            mixedSubstrateCups: mixedSubCups,
            casingCups: casingCups,
            notes: notes,
            // Optional
            writeTagTo: writeTagTo,
        }
        DoCreateRequest("fruitingChamber", body, AssertFruitingChamber, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    const batchArea = () => {
        if (substrateBatchIn) {
            return <SubstrateBatchArea readonly={true} id={substrateBatchIn._id}/>
        }
        return <SubstrateBatchSelector doSelect={(sb => {
            setSubBatch(sb)
        })} allowCreate={false/* TODO: ok?*/} creatorInPage={false/*TODO:????*/}/>
    }
    const parentArea = () => {
        if (!parent) {
            return <MainCollectionInputOrRead onIdSelected={setParentId}/>
        }
        return <div>{parentId || "unknown parent"}</div> // TODO: FIX
    }
    return (
        <NewEntryFormWrapper entryType={"fruitingChamber"}>
            <ErrorDisplay err={err}/>
            {batchArea()}
            {/* TODO: grain volume from parent????*/}
            {parentArea()}
            <div className={"inlineChildren"}>
                <div className={"text-lg"}>{"Volume of grain mixed: "}</div>
                <VolumeSelector initialVal={0} initialUnit={"quarts"} updateNumberOfCups={setVolumeGrainCups}/>
            </div>
            <div className={"inlineChildren"}>
                <div className={"text-lg"}>{"Volume of substrate mixed: "}</div>
                <VolumeSelector initialVal={0} initialUnit={"quarts"} updateNumberOfCups={setMixedSubCups}/>
            </div>
            <div className={"inlineChildren"}>
                <div className={"text-lg"}>{"Volume casing: "}</div>
                <VolumeSelector initialVal={0} initialUnit={"quarts"} updateNumberOfCups={setCasingCups}/>
            </div>
            <NewEntryNotes setNotes={setNotes}/>
            <ReaderWriterSelector onSelect={setWriteTagTo}/>
            {/* SUBMIT AREA */}
            <input type="submit" value="Submit" onClick={newFruitingChamberSubmit} onSubmit={(e) => {
                e.preventDefault();
            }}/>
        </NewEntryFormWrapper>
    )
}

export function FruitingChamberImportDisplay({headerLevel}: ImportDisplayInput) {
    const [recipe, setRecipe] = useState<SubstrateRecipeData | undefined>()
    const [creationDate, setCreationDate] = useState(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [grainCups, setGrainCups] = useState<number>(4) // TODO: set and initial?
    const [mixedSubRatio, setMixedSubRatio] = useState<number>(4) // TODO: set and initial?
    const [casingRatio, setCasingRatio] = useState<number>(2) // TODO: set and initial?
    // Non-required
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>()
    const [generation, setGeneration] = useState<number | undefined>()
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>()
    const [imageFile, setImageFile] = useState<File | undefined>()
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const cookies = useContext(CookiesContext)
    const submitImportFruitingChamber = () => {
        const reqd = new Map<string, any>([
            ['recipe', recipe],
            ['creationDate', creationDate],
            ['species', species]
        ])
        for (let [key, val] of reqd) {
            if (val === undefined) {
                setErr(key + " must be defined!");
                return
            }
        }

        let bodyObj: any = {
            recipe: recipe,
            creationDate: creationDate,
            species: species?._id,
            grainCups: grainCups,
            substrateRatio: mixedSubRatio,
            casingRatio: casingRatio,
            //perms: perms, // From spec/subspec
            // optional
            subspecies: subspecies?._id,
            generation: generation,
            knownFruitable: knownFruitable,
            writeTagTo: writeTagTo,
        }
        let formData = new FormData()
        imageFile && formData.set("img", imageFile, "img")
        setFormData(formData, bodyObj)

        MultipartImportRequest(formData, "fruitingChamber", AssertFruitingChamber, setErr, allCookies(cookies))
    }
    return <ImportEntryFormWrapper entryType={"fruitingChamber"}>
        <ErrorDisplay headerLevel={headerLevel} err={err}/>
        {/* Required Fields */}
        <SelectorWrapper current={recipe} title={"Recipe"} nameFunc={(v: SubstrateRecipeData) => v._id}>
            <SubstrateRecipeSelector doSelect={setRecipe} allowCreate={true}/>
        </SelectorWrapper>
        <DateArea pre={"Created on: "} readonly={false} when={Date.now()} updateParent={setCreationDate}/>
        <ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel}/>
        <div className={"inlineChildren"}>
            <div>{"Grain volume: "}</div>
            <VolumeSelector initialVal={1} initialUnit={"quarts"} updateNumberOfCups={setGrainCups}/>
        </div>
        <div className={"inlineChildren"}>
            <div>{"Mixed substrate ratio"}</div>
            <FloatInput initial={mixedSubRatio} onChange={setMixedSubRatio}/>
        </div>
        <div className={"inlineChildren"}>
            <div>{"Casing ratio"}</div>
            <FloatInput initial={casingRatio} onChange={setCasingRatio}/>
        </div>
        {/* Optional fields*/}
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}
                                    headerLevel={headerLevel}/>

        <GenerationInput updateParent={setGeneration}/>
        <KnownFruitableArea doSelect={setKnownFruitable} headerLevel={headerLevel}/>
        <ImageSelector updateParent={setImageFile}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        {/* SUBMIT AREA */}
        <button className={"bottomButton greenButton"} onClick={submitImportFruitingChamber}>{"Submit"}</button>
    </ImportEntryFormWrapper>
}

export function FruitingChamberListPageTable({data, onClick, withLink}: ListPageItems<FruitingChamberData>) {
    let cols: ListTableColumn<FruitingChamberData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v) => v.species || ""),
        NewColumn("Subspec", v => v.subspecies || ""),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: FruitingChamberData) => {
            return <EntryLinkWrapper
                props={{entry:v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new FruitingChamberData(v)}}/>
}

export function FruitingChamberSelectorTable({data, onClick}: ListPageItems<FruitingChamberData>) {
    return <FruitingChamberListPageTable data={data} onClick={onClick}/>
}

export function FruitingChamberSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: FruitingChamberData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: FruitingChamberData[]): JSX.Element => {
        return <FruitingChamberSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"fruitingChamber"} entryTypes={"fruitingChambers"} doSelect={doSelect}
                                   asserter={AssertFruitingChamber}
                                   table={table}>
        {allowCreate && <NewFruitingChamberForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}
